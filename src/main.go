package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"sync"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

func f(from string) {
	for i := 0; i < 3; i++ {
		fmt.Println(from, ":", i)
	}
}

func main() {

	runtime.GOMAXPROCS(2)

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Printf("hello, world\n")

	// Create a request pipeline using your Storage account's name and account key.
	// accountName, accountKey := accountInfo()
	credential, err := azblob.NewSharedKeyCredential("devstoreaccount1", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")
	if err != nil {
		log.Fatal(err)
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your Storage account blob service URL endpoint.
	cURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:10000/devstoreaccount1/fcz"))

	// Create an ServiceURL object that wraps the service URL and a request pipeline to making requests.
	containerURL := azblob.NewContainerURL(*cURL, p)

	blobsFirstSet := []string{"file1.json", "file2.json"}
	blobsSecondSet := []string{"file3.json", "file4.json"}

	ctx := context.Background() // This example uses a never-expiring context

	fmt.Println("Starting Go Routines")
	go func() {
		defer wg.Done()

		for _, blob := range blobsFirstSet {
			// go f("goroutine" + blob)
			// Here's how to create a blob with HTTP headers and metadata (I'm using the same metadata that was put on the container):
			blobURL := containerURL.NewBlockBlobURL(blob)

			// Here's how to read the blob's data with progress reporting:
			get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)

			if err != nil {
				log.Fatal(err)
			}

			// Wrap the response body in a ResponseBodyProgress and pass a callback function for progress reporting.
			responseBody := pipeline.NewResponseBodyProgress(get.Body(azblob.RetryReaderOptions{}),
				func(bytesTransferred int64) {
					fmt.Printf("[Routine 1] - Read %d of %d bytes from blob %s.\n", bytesTransferred, get.ContentLength(), blob)
				})

			downloadedData := &bytes.Buffer{}
			downloadedData.ReadFrom(responseBody)
			responseBody.Close() // The client must close the response body when finished with it
			// The downloaded blob data is in downloadData's buffer
		}

	}()

	go func() {
		defer wg.Done()

		for _, blob := range blobsSecondSet {
			// go f("goroutine" + blob)
			// Here's how to create a blob with HTTP headers and metadata (I'm using the same metadata that was put on the container):
			blobURL := containerURL.NewBlockBlobURL(blob)

			// Here's how to read the blob's data with progress reporting:
			get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)

			if err != nil {
				log.Fatal(err)
			}

			// Wrap the response body in a ResponseBodyProgress and pass a callback function for progress reporting.
			responseBody := pipeline.NewResponseBodyProgress(get.Body(azblob.RetryReaderOptions{}),
				func(bytesTransferred int64) {
					fmt.Printf("[Routine 2] - Read %d of %d bytes from blob %s.\n", bytesTransferred, get.ContentLength(), blob)
				})

			downloadedData := &bytes.Buffer{}
			downloadedData.ReadFrom(responseBody)
			responseBody.Close() // The client must close the response body when finished with it
			// The downloaded blob data is in downloadData's buffer
		}
	}()

	fmt.Println("Waiting To Finish")
	wg.Wait()

	fmt.Println("\nTerminating Program")

}
