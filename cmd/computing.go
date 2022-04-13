package cmd

import (
	"fmt"
	"io/ioutil"

	"apirecorder/nifcloud"

	"github.com/spf13/cobra"
)

// computingCmd represents the computing command
var computingCmd = &cobra.Command{
	Use:   "computing",
	Short: "computingAPI用リクエストURL生成",
	Long: `request.json に定義したリクエスト内容を基にリクエスト用URLを生成します.
	
第一引数にリージョンを指定することが可能です.

ex. go run main.go computing jp-east-1`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("computing called")
		bytes, err := ioutil.ReadFile("request.json")
		if err != nil {
			panic(err)
		}
		fmt.Println("↓リクエスト用URL↓")
		fmt.Println(nifcloud.CreateRequest(args[0], string(bytes)))
	},
}

func init() {
	rootCmd.AddCommand(computingCmd)
}
