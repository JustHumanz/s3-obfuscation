package pkg

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

// Decode the filename/dir from base64 to string
func DecodeBase64(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// If decoding fails, return the original string
		return encoded
	}
	return string(decoded)
}

// Encode the filename/dir from string to base64
func EncodeBase64(input string) string {
	data := []byte(input)
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded
}

// Function to compute the MD5 hash of a string
func ComputeMD5(input string) string {
	var md5Hash = md5.New()

	_, err := md5Hash.Write([]byte(input))
	if err != nil {
		log.Fatalln(err)
	}

	hashInBytes := md5Hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

// Create a temporary file in /tmp
func CreateTempFile() (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	return tmpFile, nil
}

// Print the index
func PrintIndex(data map[string]interface{}, path string) {
	for _, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			pathName := path + "/" + DecodeBase64(v["name"].(string))
			if v["type"] == "dir" {
				PrintIndex(v["sub"].(map[string]interface{}), pathName)
			} else if v["type"] == "file" {
				fmt.Println(pathName)
			}
		default:
			log.Fatal("Invalid json")
		}
	}
}

// Update the index file when add/delete the obj
func UpdateDataIndex(data map[string]interface{}, NewDataStruct UpdateIndex, add bool) {
	for v := range NewDataStruct.Index {
		NewIndex := NewDataStruct.Index[v]
		NewMap := NewDataStruct.Map[NewDataStruct.Index[v]]

		var findAndUpdateIndex func(map[string]interface{}) map[string]interface{}
		findAndUpdateIndex = func(data2 map[string]interface{}) map[string]interface{} {

			if data2[NewIndex] != nil {
				typeFile := data2[NewIndex].(map[string]interface{})["type"]
				if typeFile == "dir" {
					sub := data2[NewIndex].(map[string]interface{})["sub"]
					return sub.(map[string]interface{})
				} else {
					if !add {
						delete(data2, NewIndex)
					}

					return data2
				}
			} else {
				if add {
					data2[NewIndex] = make(map[string]interface{})
					data2[NewIndex] = NewMap
				}

				typeFile := data2[NewIndex].(map[string]interface{})["type"]
				if typeFile == "dir" {
					sub := data2[NewIndex].(map[string]interface{})["sub"]
					return sub.(map[string]interface{})
				} else {
					return data2
				}
			}
		}
		data = findAndUpdateIndex(data)
	}
}

// Read the index as json
func ReadJSON(filename string, target interface{}) error {
	// Read the file content
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	if e := reflect.ValueOf(fileData); e.IsNil() {
		target = make(map[string]interface{})
	} else {
		err = json.Unmarshal(fileData, target)
		if err != nil {
			return fmt.Errorf("error unmarshaling JSON: %v", err)
		}

		fmt.Printf("Data successfully read from %s\n", filename)
	}

	return nil
}

// Save the index in json file
func SaveJson(data map[string]interface{}, filename string) error {
	// Convert the map to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "    ") // Indent with 4 spaces
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Write the JSON data to the file
	err = os.WriteFile(filename, jsonData, 0644) // 0644 is the file permission
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Printf("Data successfully saved to %s\n", filename)
	return nil
}
