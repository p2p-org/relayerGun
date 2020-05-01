package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/iqlusioninc/relayer/relayer"
	"github.com/spf13/cobra"
)

func testnetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "testnets",
		Aliases: []string{"tst"},
		Short:   "commands for joining and running relayer testnets",
	}
	cmd.AddCommand(
		faucetStartCmd(),
		faucetRequestCmd(),
		faucetMilkCmd(),
	)
	return cmd
}

func faucetRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "request [chain-id] [[key-name]]",
		Aliases: []string{"req"},
		Short:   "request tokens from a relayer faucet",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := config.Chains.Get(args[0])
			if err != nil {
				return err
			}

			done := chain.UseSDKContext()
			defer done()

			urlString, err := cmd.Flags().GetString(flagURL)
			if err != nil {
				return err
			}

			if urlString == "" {
				u, err := url.Parse(chain.RPCAddr)
				if err != nil {
					return err
				}

				host, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					return err
				}

				urlString = fmt.Sprintf("%s://%s:%d", u.Scheme, host, 8000)
			}

			var keyName string
			if len(args) == 2 {
				keyName = args[1]
			} else {
				keyName = chain.Key
			}

			info, err := chain.Keybase.Key(keyName)
			if err != nil {
				return err
			}

			body, err := json.Marshal(relayer.FaucetRequest{Address: info.GetAddress().String(), ChainID: chain.ChainID})
			if err != nil {
				return err
			}

			resp, err := http.Post(urlString, "application/json", bytes.NewBuffer(body))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(respBody))
			return nil
		},
	}
	return urlFlag(cmd)
}

func faucetStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faucet [chain-id] [key-name] [amount]",
		Short: "listens on a port for requests for tokens",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := config.Chains.Get(args[0])
			if err != nil {
				return err
			}
			info, err := chain.Keybase.Key(args[1])
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoin(args[2])
			if err != nil {
				return err
			}
			listenAddr, err := cmd.Flags().GetString(flagListenAddr)
			if err != nil {
				return err
			}
			r := mux.NewRouter()
			r.HandleFunc("/", chain.FaucetHandler(info.GetAddress(), amount)).Methods("POST")
			srv := &http.Server{
				Handler:      r,
				Addr:         listenAddr,
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  15 * time.Second,
			}
			chain.Log(fmt.Sprintf("Listening on %s for faucet requests...", listenAddr))
			return srv.ListenAndServe()
		},
	}
	return listenFlag(cmd)
}

func faucetMilkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "milk [chain-id] [[workers-count]]",
		Aliases: []string{"req"},
		Short:   "milk the faucet",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := config.Chains.Get(args[0])
			if err != nil {
				return err
			}

			urlString, err := cmd.Flags().GetString(flagURL)
			if err != nil {
				return err
			}

			if urlString == "" {
				u, err := url.Parse(chain.RPCAddr)
				if err != nil {
					return err
				}

				host, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					return err
				}

				urlString = fmt.Sprintf("%s://%s:%d", u.Scheme, host, 8000)
			}

			workersCount := 1

			if len(args) > 1 {
				workersCount, err = strconv.Atoi(args[1])
				if err != nil {
					return err
				}
			}

			faucetTimeout := 5 * time.Minute

			addressLastRequest := make(map[string]time.Time, len(chain.Keys))
			addresses := make([]string, len(chain.Keys))

			log.Println("Init keys...")
			for _, key := range chain.Keys {
				info, err := chain.Keybase.Key(key)
				if err != nil {
					return err
				}
				address := info.GetAddress().String()
				addressLastRequest[address] = time.Unix(0, 0)
				addresses = append(addresses, address)
			}
			log.Println("Keys initialized!")

			milker := InitFaucetMilker(workersCount, addressLastRequest, addresses, urlString, chain.ChainID, faucetTimeout)

			milker.Start()

			return nil
		},
	}
	return urlFlag(cmd)
}

type FaucetMilker struct {
	sync.Mutex
	addressLastRequest map[string]time.Time
	addresses          []string
	workersCount       int
	jobs               chan relayer.FaucetRequest
	chainID            string
	url                string
	faucetTimeout      time.Duration
}

func InitFaucetMilker(workersCount int, addressLastRequest map[string]time.Time, addresses []string, url, chainID string,
	faucetTimeout time.Duration) *FaucetMilker {
	return &FaucetMilker{
		addressLastRequest: addressLastRequest,
		workersCount:       workersCount,
		jobs:               make(chan relayer.FaucetRequest, len(addressLastRequest)),
		url:                url,
		chainID:            chainID,
		faucetTimeout:      faucetTimeout,
		addresses:          addresses,
	}
}

func (f *FaucetMilker) Start() {
	for w := 0; w < f.workersCount; w++ {
		go f.worker()
	}

	for {
		for _, address := range f.addresses {
			f.Lock()
			lastRequestTime := f.addressLastRequest[address]
			f.Unlock()
			if time.Since(lastRequestTime) < f.faucetTimeout {
				continue
			}
			f.jobs <- relayer.FaucetRequest{Address: address, ChainID: f.chainID}
		}
	}
}

func (f *FaucetMilker) worker() {
	for j := range f.jobs {
		body, err := json.Marshal(j)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		log.Println("Sending request...")
		resp, err := http.Post(f.url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Println(err.Error())
			continue
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			log.Println(err.Error())
		}
		resp.Body.Close()

		log.Println(string(respBody))

		f.Lock()
		f.addressLastRequest[j.Address] = time.Now()
		f.Unlock()
	}
}
