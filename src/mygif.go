package main

import (
	"image/gif"
	"log"
	"os"
    "fmt"

)

func main() {
	cloneImage("shapes")
	//process("blob")
}

func cloneImage(filename string){
    f, err := os.Open(filename + ".gif")
  
	if err != nil {
		log.Fatal(err.Error())
	}

    im, err := gif.DecodeAll(f)
     fmt.Println( "width:",im.Config.Width)
     fmt.Println( "height:",im.Config.Height)
     fmt.Println("images",len(im.Image))
     fmt.Println("count",im.LoopCount)

    out, err := os.Create(filename + ".test.gif")

    gif.EncodeAll(out,im)

	defer f.Close()
    defer out.Close()
}