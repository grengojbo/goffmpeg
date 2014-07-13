package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	// "strings"
	"syscall"

	"github.com/grengojbo/beego/modules/utils"

	"github.com/grengojbo/goffmpeg"

	"github.com/codegangsta/cli"

	// "github.com/bitly/go-nsq"
	// "github.com/bitly/nsq/util"
)

type TranscodeHandler struct {
	totalMessages int
	messagesShown int
}

func Contains(slice interface{}, val interface{}) bool {
	sv := reflect.ValueOf(slice)

	for i := 0; i < sv.Len(); i++ {
		if sv.Index(i).Interface() == val {
			return true
		}
	}
	return false
}

func CompareArrays(a, b []string) (c []string) {
	for _, i := range a {
		if !Contains(b, i) {
			// fmt.Println(">", i)
			c = append(c, i)
		}
	}
	return c
}

func main() {

	// var tasks = []string{"cook", "clean", "laundry", "eat", "sleep", "code"}
	var movieList = []string{"de_glaimond.mov", "FS-GIAMBATTISTA_VALLI-AW14-15-040414.mpg", "Kropp.mov", "aabb_inside.avi"}
	// var movieList = []string{"05min_FS-JEAN_PAUL_GAULTIER_SS14-030414.mpg", "de_glaimond.mov", "FS-GIAMBATTISTA_VALLI-AW14-15-040414.mpg", "Kropp.mov", "aabb_inside.avi"}
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "tv-probe"
	app.Usage = "make an explosive entrance"

	app.Flags = []cli.Flag{
		cli.StringFlag{"lang, l", "english", "language for the greeting"},
		cli.StringFlag{"to", "h720p", "Transcode to 720p"},
		// cli.BoolFlag{"deamon, d", false, "Run as Daemon"},
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	tr := goffmpeg.Transcoder{DirVideoVod: "/Users/jbo/Movies/vod/", DirConvTs: "/Users/jbo/Movies/convts/", DirTs: "/Users/jbo/Movies/channel1/", DirTmp: "/Users/jbo/Movies/tmp/"}
	tr.Init()
	// tr := goffmpeg.Transcoder{}
	// tr.dirVideoVod = "/opt/vod/"
	// tr.dirConvTs = "/Users/jbo/Movies/convts/"
	// tr.dirTs = "/opt/channel1/"

	app.Commands = []cli.Command{
		{
			Name:      "transcode",
			ShortName: "t",
			Usage:     "Transcode",
			Action: func(c *cli.Context) {
				if c.GlobalString("lang") == "uk" {
					fmt.Println("lang Ok!")
				}
				if len(c.Args()) > 0 {
					// movie := tr.SetMovie(c.Args().First())
					// _, filename := filepath.Split(movie.Format.Filename)
					dir, filename := filepath.Split(c.Args().First())
					if movie, err := tr.FfProbe(dir+"/", filename); err != nil {
						log.Fatal(err)
					} else {
						fmt.Println("Convert: ", filename)
						// fmt.Println(movie)
						// movie.Show()
						// tr.SetStream(movie)
						movie.SetStream(&tr)
						tr.BitRate = "3000k"
						// tr.DurationSeonds, tr.Duration = utils.DurationToTime("2")

						if fout, err := tr.ConvMp4("hd720", tr.DurationSeonds); err != nil {
							log.Fatal(err)
						} else {
							fmt.Println("Movies", fout)
							if ts, err := tr.BuildTs(fout, "hd720"); err != nil {
								log.Fatal(err)
							} else {
								fmt.Println("TS", ts)
							}
						}
					}

					// fileTs := strings.TrimSpace(strings.ToLower(strings.Replace(filename, filepath.Ext(filename), ".ts", 1)))
					// movie.Show()
					// fmt.Println("Convert to: ", movie.Format.Filename, " ts:", fileTs)
					// _ = tr.Conv("720p", movie.Format.Filename, filename, "mp4")
					// _ = this.BuildTs(fn, fts)
				} else {
					fmt.Println("Convert Multy")
				}
				// convertMovie := goffmpeg.ListsMovie("/Users/jbo/Movies/convts/")
				// fmt.Println("convert: ", convertMovie)
				os.Exit(0)
			},
		},
		{
			Name:      "info",
			ShortName: "i",
			Usage:     "Информация о Видео",
			Action: func(c *cli.Context) {
				if c.GlobalString("lang") == "uk" {
					fmt.Println("lang Ok!")
				}
				if err := utils.IsExits(c.Args().First()); err == nil {
					dir, filename := filepath.Split(c.Args().First())
					fmt.Println("Info : ", c.Args().First(), " dir: ", dir, filename)
					if movie, err := tr.FfProbe(dir+"/", filename); err != nil {
						log.Fatal(err)
					} else {
						fmt.Println("Convert: ", filename)
						// fmt.Println(movie)
						// movie.Show()
						tr.SetStream(movie)
					}
				} else {
					log.Fatal(err.Error())
				}
				// convertMovie := goffmpeg.ListsMovie("/Users/jbo/Movies/convts/")
				// fmt.Println("convert: ", convertMovie)
				fmt.Println("----------------------------------------------------")
				os.Exit(0)
			},
		},
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "List dirs",
			Action: func(c *cli.Context) {
				convertMovie := tr.ListCovert()
				// fmt.Println("convert: ", convertMovie)
				// fmt.Println("movieList", movieList)
				newConv := CompareArrays(convertMovie, movieList)
				for _, i := range newConv {
					if movie, err := tr.FfProbe(tr.DirConvTs, i); err != nil {
						log.Fatal(err)
					} else {
						fmt.Println("Convert: ", i)
						// fmt.Println(movie)
						movie.Show()
					}
					// goffmpeg.PeLs(tr.DirConvTs, i)
				}
				os.Exit(0)
			},
		},
	}

	app.Run(os.Args)
	// for {
	//   select {
	//   case <-r.StopChan:
	//     return
	//   case <-sigChan:
	//     r.Stop()
	//   }
	// }
	<-sigChan

}
