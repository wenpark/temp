// Debian: sudo apt-get install libmagickwand-dev
package main

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
    "image"
	"image/color/palette"
	"image/draw"
    "image/jpeg"
	"image/png"
	"image/gif"
    "github.com/nfnt/resize"
)

var destDir string

func init() {
	flag.StringVar(&destDir, "p", "", "destination dir path to watch")
}

func isDirectory(fpath string) bool {
	fileInfo, err := os.Stat(fpath)

	if err != nil {
		panic(err)
	}

	if fileInfo.IsDir() {
		return true
	}

	return false
}

func genImageFiles(imagePath string) error {

	fileExt := filepath.Ext(imagePath)
	name := strings.TrimSuffix(imagePath, fileExt)

	image240Path := name + "_240" + fileExt
	image480Path := name + "_480" + fileExt
	image720Path := name + "_720" + fileExt

	image240, err := os.OpenFile(image240Path, os.O_RDWR|os.O_CREATE, 0666)

	defer image240.Close()

	if err != nil {
		panic(err)
	}

	image480, err := os.OpenFile(image480Path, os.O_RDWR|os.O_CREATE, 0666)

	defer image480.Close()

	if err != nil {
		panic(err)
	}

	image720, err := os.OpenFile(image720Path, os.O_RDWR|os.O_CREATE, 0666)

	defer image720.Close()

	if err != nil {
		panic(err)
	}

    log.Print("Processing", imagePath)
    originImage, err := os.Open(imagePath)

    img,imgFmt,err := image.Decode(originImage)

    defer image240.Close()

    m240 := resize.Resize(240, 0, img, resize.Bilinear)
    m480 := resize.Resize(480, 0, img, resize.Bilinear)
    m720 := resize.Resize(720, 0, img, resize.Bilinear)

    if imgFmt == "jpeg" {
		jpeg.Encode(image240, m240,nil)
        jpeg.Encode(image480, m480,nil)
        jpeg.Encode(image720, m720,nil)
	}else if imgFmt == "png" {
		png.Encode(image240, m240)
        png.Encode(image480, m480)
        png.Encode(image720, m720)
	}else if imgFmt == "gif" {
        processGif(imagePath,image240,240)
        processGif(imagePath,image480,480)
        processGif(imagePath,image720,720)
	} else {
        log.Print("Unrecognized format:", imgFmt)
	}
    

	return nil
}

func processGif(originImage string,targetFile *os.File,width uint){
    	f, err := os.Open(originImage)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()
   
    im, err := gif.DecodeAll(f)
    
    if err != nil {
		log.Fatal(err.Error())
	}

    firstFrame := im.Image[0].Bounds()
	b := image.Rect(0, 0, firstFrame.Dx(), firstFrame.Dy())
	img := image.NewRGBA(b)

    // Resize each frame.
	for index, frame := range im.Image {
		bounds := frame.Bounds()
		draw.Draw(img, bounds, frame, bounds.Min, draw.Over)
		im.Image[index] = gifImageToPaletted(ProcessGifImage(width,img))
	}
    im.Config = image.Config{}
    fileInfo,_ :=  targetFile.Stat()
     log.Print("-----create gif------",fileInfo.Name())
    gif.EncodeAll(targetFile, im)

}

func ProcessGifImage(width uint,img image.Image) image.Image {
	return resize.Resize(width, 0, img, resize.NearestNeighbor)
}

func gifImageToPaletted(img image.Image) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, palette.Plan9)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}

func isGenFile(genImagePath string) bool {
	if strings.LastIndex(genImagePath, "_240") > -1 || strings.LastIndex(genImagePath, "_480") > -1 || strings.LastIndex(genImagePath, "_720") > -1 {
		return true
	}
	return false
}

func isHiddenFile(genImagePath string) bool {
	if strings.LastIndex(filepath.Base(genImagePath), ".") == 0 {
		return true
	}
	return false
}

//EOF error ad hoc
func isImageType(imagePath string) (bool, error) {
	file, err := os.Open(imagePath)
	defer file.Close()
	buff := make([]byte, 512)

	_, err = file.Read(buff)
	if err != nil {

		return false, err
	}

	filetype := http.DetectContentType(buff)

	switch filetype {
	case "image/jpeg", "image/jpg":
		return true, err
	case "image/gif":
		return true, err
	case "image/png":
		return true, err
	case "application/pdf":
		return false, err
	default:
		return false, err

	}

}

func main() {
	flag.Parse()

	if !isDirectory(destDir) {
		panic("Is not DIR")
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Print("--------event--------", event.String())
				if event.Op == fsnotify.Create  {
					//	log.Println("creates file:", event.Name)

					if isDirectory(event.Name) {
						log.Println("watch directory!", event.Name)
						watcher.Add(event.Name)
						continue
					}

					if isHiddenFile(event.Name) {
						continue
					}

					if isGenFile(event.Name) {
						continue
					}

					isImage, err := isImageType(event.Name)

					if isImage {
						// can not use goroutine
						genImageFiles(event.Name)
					} else {
						if err != nil {
							log.Println(err)
						}
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	//	err = watcher.Add(destDir)

	err = filepath.Walk(destDir, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			watcher.Add(walkPath)
			log.Println("add watch dir : ", walkPath)
		}
		return nil

	})

	if err != nil {
		log.Fatal(err)
	}
	<-done

}
