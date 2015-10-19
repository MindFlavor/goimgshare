package thumb

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mindflavor/goimgshare/folders/physical"
	"github.com/nfnt/resize"
)

// Cache is the thumbnail
// cache type.
type Cache struct {
	cachePath string
	maxWidth  uint
	maxHeight uint
}

// New initialized a new Cache with
// specified parameters.
func New(cachePath string, maxWidth, maxHeight uint) *Cache {
	return &Cache{cachePath, maxWidth, maxHeight}
}

// GetThumb returns the thumbnail
// stream or an error.
func (cache *Cache) GetThumb(folder *physical.Folder, name string) (io.Reader, error) {
	if err := cache.generateJpegThumbnail(folder, name); err != nil {
		return nil, err
	}

	folderPath, err := cache.folderPath(folder)
	if err != nil {
		return nil, err
	}
	cacheImgFullPath := filepath.Join(folderPath, name)

	file, err := os.Open(cacheImgFullPath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	// populate a bytes buffer so we can close the file handle at
	// the end of the function.
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	return buf, nil
}

func (cache *Cache) generateJpegThumbnail(folder *physical.Folder, name string) error {
	// check for file existence
	folderPath, err := cache.folderPath(folder)
	if err != nil {
		return err
	}

	cacheImgFullPath := filepath.Join(folderPath, name)
	if _, err := os.Stat(cacheImgFullPath); err != nil {
		// if there is no thumb, generate it first
		if os.IsNotExist(err) {
			srcImage := filepath.Join(folder.Path, name)
			// generate thumb
			file, err := os.Open(srcImage)
			if err != nil {
				log.Printf("ERROR: thumbcache::generateJpegThumbnail - Cannot open file: %q", err)
				return err
			}
			defer file.Close()

			var img image.Image
			err = nil

			ext := strings.ToLower(filepath.Ext(srcImage))
			switch ext {
			case ".jpg":
				img, err = jpeg.Decode(file)
			case ".jpeg":
				img, err = jpeg.Decode(file)
			case ".gif":
				img, err = gif.Decode(file)
			case ".png":
				img, err = png.Decode(file)
			}

			if err != nil {
				log.Printf("ERROR: thumbcache::generateJpegThumbnail - Cannot decode file: %q", err)
				return err
			}

			imgThumb := resize.Thumbnail(cache.maxWidth, cache.maxHeight, img, resize.Lanczos3)

			// create thumb
			imgOut, err := os.Create(cacheImgFullPath)

			if err != nil {
				log.Printf("ERROR: thumbcache::generateJpegThumbnail - Cannot create thumb file: %q", err)
				return err
			}
			defer imgOut.Close()

			if err := jpeg.Encode(imgOut, imgThumb, nil); err != nil {
				return err
			}

			return nil
		}

		// if is not FileNotExists error we bail :)
		return err
	}

	return nil
}

func (cache *Cache) folderPath(folder *physical.Folder) (string, error) {
	sResolution := fmt.Sprintf("%dx%d", cache.maxWidth, cache.maxHeight)

	folderPath := filepath.Join(cache.cachePath, sResolution, folder.ID)
	if err := os.MkdirAll(folderPath, 07777); err != nil {
		return "", err
	}

	return folderPath, nil
}
