package main

import (
	"errors"
	"fmt"
	"fuzz_debug_platform/sqlfuzz"
	"fuzz_debug_platform/view"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var yyPath string
var port int
var db string
var dsn1 string
var dsn2 string
var debug bool
var queries int

var rootCmd = &cobra.Command{
	Use:   "sql fuzz debug platform",
	Short: "tidb sql fuzz and debug platform",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if yyPath == "" {
			return errors.New("yy are required")
		}

		if dsn1 == "" || dsn2 == "" {
			return errors.New("db dsn1 or dsn2 are all required")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		yyBytes, err := ioutil.ReadFile(yyPath)
		if err != nil {
			log.Fatalf("read yy path fail, %v\n", err)
		}

		yyContent := string(yyBytes)
		go func() {
			sqlfuzz.Fuzz(yyContent, dsn1, dsn2, queries, debug)
		}()

		httpHandle("/graph", view.Graph(yyContent))
		httpHandle("/heat", view.Heat())
		httpHandle("/codepos", view.CodePos())

		log.Printf("listen on :%d\n", port)
		log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	},
}

func init() {
	rootCmd.Flags().StringVarP(&yyPath, "yy", "Y", "", "file path of randgen yy file")
	rootCmd.Flags().IntVarP(&port, "port", "P", 43000, "the port to listen")
	rootCmd.Flags().StringVar(&dsn1, "dsn1", "", "db to test")
	rootCmd.Flags().StringVar(&dsn2, "dsn2", "", "standard db")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "open sql debug")
	rootCmd.Flags().IntVarP(&queries, "queries", "Q", 1000, "queries to generate")
}

func httpHandle(path string, handler http.HandlerFunc) {
	http.HandleFunc(path, crossDomain(handler))
}

func crossDomain(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		writer.Header().Set("content-type", "application/json")
		handler(writer, request)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(rootCmd.UsageString())
		os.Exit(1)
	}
}
