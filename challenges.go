package main

import (
  "fmt"
  "net/http"
  "log"
  "bufio"
  "encoding/base64"
  "os"
  "strings"
  "time"
)

func createTestHandler(w http.ResponseWriter, r *http.Request) {
  //fmt.Fprintf(w, "", r.URL.Path[1:])

  // Take details from query string
  query := r.URL.Query()
  candidate := query["candidate"][0]
  fileLocation := query["fileLocation"][0]

  if candidate == "" || fileLocation == "" {
    log.Fatal ("No candidate or no file location")
    return
  }

  // Generate id
  salt := "stratus5"
  saltedData := []byte(candidate+salt)
  id := base64.StdEncoding.EncodeToString(saltedData)
  
  addLineToFile("undownloaded.csv.txt", []string{id, candidate, fileLocation})

  
  // Return url of download
  fmt.Fprintf(w, "/downloadTest/%v", id)
}

func downloadTestHandler(w http.ResponseWriter, r *http.Request) {

  // Retrieve id from query
  qId := r.URL.Path[14:]

  var result []string
  var misses [][]string

  // Read undownloaded tests file into memory
  f, err := os.Open("undownloaded.csv.txt")
  if err != nil {
    log.Fatal(err)
    return
  }
  
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    line := scanner.Text()    
    attrs := strings.Split(line, ",")
    id := attrs[0]    
    if id == qId {
      result = attrs
      continue
    }
    
    misses = append(misses, attrs)
    
  }  

  // If not found, return 404
  if result == nil {
    http.Error(w, "test not found", 404)
    return
  }

  // Add date to result
  result = append(result, time.Now().String())

  // Write to downloaded file [id, candidate, fileLocation, downloaded date]
  addLineToFile("downloaded.csv.txt", result)
  
  // Rewrite undownloaded file minus entry
  writeToFile("undownloaded.csv.txt", misses)
  
  // Load file, and send to customer
  http.ServeFile(w, r, result[2])  
  
}

func downloadedHandler(w http.ResponseWriter, r *http.Request) {  
  http.ServeFile(w, r, "downloaded.csv.txt")  
}

func undownloadedHandler(w http.ResponseWriter, r *http.Request) {  
  http.ServeFile(w, r, "undownloaded.csv.txt")  
}

func reset(w http.ResponseWriter, r *http.Request){
  files := []string{"downloaded.csv.txt", "undownloaded.csv.txt"}
  for _, fileName := range files {
    fo, err := os.Create(fileName)
    if err != nil { 
      log.Fatal(err)
      return
    }
    defer fo.Close()
  }
  fmt.Fprint(w, "Files reset")
}

func addLineToFile(fileName string, attrs []string){
  // Add line in test file
  fo, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND, 0666)
  if err != nil { 
    log.Fatal(err)
    return
  }
  defer fo.Close()
  
  fw := bufio.NewWriter(fo)
  lastEntryIndex := len(attrs) -1
  for i,att := range attrs {        
    fmt.Fprintf(fw, "%v", att)
    // if not the last attribute add a comma
    if i < lastEntryIndex {
      fmt.Fprint(fw, ",")
    }
    // if the last attribute add a new line
    if i == lastEntryIndex {
      fmt.Fprint(fw, "\n")
    }
  }
  fw.Flush()
}

func writeToFile(fileName string, lines [][]string){
  fo, err := os.Create(fileName)
  if err != nil { 
    log.Fatal(err)
    return
  }
  defer fo.Close()
  
  fw := bufio.NewWriter(fo)
  for i := 0; i<len(lines); i++ {
    line := lines[i]
    lineLength := len(line)
    for j := 0; j < lineLength; j++ { 
      fmt.Fprintf(fw, "%v", line[j])
      // if not the last attribute add a comma
      if j < lineLength -1 {
        fmt.Fprint(fw, ",")
      }
      // if the last attribute add a new line
      if j == lineLength - 1 {
        fmt.Fprint(fw, "\n")
      }
    }
  }
  fw.Flush()
}

func main(){
  http.HandleFunc("/createTest", createTestHandler)
  http.HandleFunc("/downloaded", downloadedHandler)
  http.HandleFunc("/undownloaded", undownloadedHandler)
  http.HandleFunc("/downloadTest/", downloadTestHandler)
  http.HandleFunc("/reset", reset)
  http.ListenAndServe(":8080", nil)
}
