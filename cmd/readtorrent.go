package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/dpnam2112/bittorrent-client/torrent"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read a torrent file",
	Long:  `Reads the content of a torrent file and displays it.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("Missing file path. Usage: program read ./path/to/file.torrent")
			return
		}

		filePath := args[0]
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
			return
		}

		torrent, err := torrent.ParseTorrent(file)

		if err != nil {
			log.Println("Error parsing torrent file:", err)
			return
		}

		fmt.Println(torrent.String())
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
