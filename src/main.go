package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"sync"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gorilla/mux"
)

type User struct {
	ID        string `json:"id,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
}

var users []User

// Create a request pipeline using your Storage account's name and account key.
// accountName, accountKey := accountInfo()
var credential, err = azblob.NewSharedKeyCredential("devstoreaccount1", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")

//TODO
// if err != nil {
// 	log.Fatal(err)
// }
var p = azblob.NewPipeline(credential, azblob.PipelineOptions{})

// From the Azure portal, get your Storage account blob service URL endpoint.
var cURL, _ = url.Parse(fmt.Sprintf("http://127.0.0.1:10000/devstoreaccount1/fcz"))

// Create an ServiceURL object that wraps the service URL and a request pipeline to making requests.
var containerURL = azblob.NewContainerURL(*cURL, p)
var ctx = context.Background() // This example uses a never-expiring context

func process(routineId int, blobURL azblob.BlockBlobURL) {

	// Here's how to read the blob's data with progress reporting:
	get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)

	if err != nil {
		log.Fatal(err)
	}

	// Wrap the response body in a ResponseBodyProgress and pass a callback function for progress reporting.
	responseBody := pipeline.NewResponseBodyProgress(get.Body(azblob.RetryReaderOptions{}),
		func(bytesTransferred int64) {
			fmt.Printf("[Routine %d] - Read %d of %d bytes from blob %s.\n", routineId, bytesTransferred, get.ContentLength(), blobURL.BlobURL)
		})

	downloadedData := &bytes.Buffer{}
	downloadedData.ReadFrom(responseBody)
	responseBody.Close() // The client must close the response body when finished with it
	// The downloaded blob data is in downloadData's buffer
}

func Run(w http.ResponseWriter, req *http.Request) {

	var wg sync.WaitGroup
	wg.Add(2)

	blobsFirstSet := []string{"file1.json", "file2.json"}
	blobsSecondSet := []string{"file3.json", "file4.json"}

	fmt.Println("Starting Go Routines")
	go func() {
		defer wg.Done()

		for _, blob := range blobsFirstSet {
			blobURL := containerURL.NewBlockBlobURL(blob)
			process(1, blobURL)
		}

	}()

	go func() {
		defer wg.Done()

		for _, blob := range blobsSecondSet {
			blobURL := containerURL.NewBlockBlobURL(blob)
			process(2, blobURL)
		}
	}()

	fmt.Println("Waiting To Finish")
	wg.Wait()

	fmt.Println("\nTerminating Program")

	w.Write([]byte("OK")) // Write response body
}

func main() {

	runtime.GOMAXPROCS(2)

	fmt.Println("magic is happening on port 8000")

	//creating local array
	users = append(users, User{ID: "1", Firstname: "Swanand", Lastname: "Keskar"})
	users = append(users, User{ID: "2", Firstname: "Akshay", Lastname: "Joshi"})

	router := mux.NewRouter()
	router.HandleFunc("/run", Run).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))

}
