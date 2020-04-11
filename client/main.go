package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hemanrnjn/grpc-stream/proto"
	"google.golang.org/grpc"
)

type resp map[string]interface{}

var client proto.StreamServiceClient

func main() {

	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err.Error())
	}

	client = proto.NewStreamServiceClient(conn)

	http.Handle("/", handlers(client))
	http.ListenAndServe(":8000", nil)
}

func handlers(client proto.StreamServiceClient) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", indexPage).Methods("GET")
	router.HandleFunc("/media/{mID:[0-9]+}/stream/", streamHandler).Methods("GET")
	router.HandleFunc("/media/{mID:[0-9]+}/stream/{segName:index[0-9]+.ts}", streamHandler).Methods("GET")
	return router
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func streamHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	// TODO: To be done later for multiple media files
	// mID, err := strconv.Atoi(vars["mID"])
	// if err != nil {
	// 	fmt.Println("ERROR FOUND")
	// 	response.WriteHeader(http.StatusNotFound)
	// 	return
	// }
	// fmt.Println("mID: ", mID)

	segName, ok := vars["segName"]
	if !ok {
		serveHlsM3u8(response, request, "../media", "index.m3u8")
	} else {
		serveHlsTs(response, request, "../media", segName)
	}
}

func serveHlsM3u8(w http.ResponseWriter, r *http.Request, mediaBase, m3u8Name string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, m3u8Name)
	req := &proto.Request{Filename: mediaFile}

	if response, err := client.GetFile(context.Background(), req); err == nil {
		fo, err := os.Create(mediaFile)
		if err != nil {
			log.Fatal("Failed to Create File")
		}
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
		if _, err := fo.Write(response.GetContent()); err != nil {
			log.Fatal("Error in writing file: ", err.Error())
		}
		http.ServeFile(w, r, mediaFile)
		w.Header().Set("Content-Type", "application/x-mpegURL")
	} else {
		fmt.Println("Error getting response: ", err.Error())
	}
}

func serveHlsTs(w http.ResponseWriter, r *http.Request, mediaBase, segName string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, segName)
	req := &proto.Request{Filename: mediaFile}

	if response, err := client.GetFile(context.Background(), req); err == nil {
		fo, err := os.Create(mediaFile)
		if err != nil {
			log.Fatal("Failed to Create File")
		}
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
		if _, err := fo.Write(response.GetContent()); err != nil {
			log.Fatal("Error in writing file: ", err.Error())
		}
		http.ServeFile(w, r, mediaFile)
		w.Header().Set("Content-Type", "video/MP2T")
	} else {
		fmt.Println("Error getting response: ", err.Error())
	}
}
