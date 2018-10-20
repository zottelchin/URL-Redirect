package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := bolt.Open("bolt.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("buck"))
		if err != nil {
			return fmt.Errorf("Fehler beim Bucket erstellen")
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	r := gin.Default() //TODO: USE REST instead of Path
	r.GET("/i/:value", func(c *gin.Context) {
		value := c.Param("value")
		key := put(value, db)
		c.JSON(200, gin.H{
			"key": key,
		})
	})
	r.GET("/e/:key", func(c *gin.Context) {
		key := c.Param("key")
		value := get(key, db)
		if value != "" {
			c.Redirect(302, "http://"+value)
		}
		c.JSON(200, gin.H{
			"message": value,
		})
	}) //TODO: Redirect to create new Path if key is not in database
	r.GET("/d/:key", func(c *gin.Context) {
		key := c.Param("key")
		value := get(key, db)
		if value != "" {
			remove(key, db)
			c.JSON(200, gin.H{
				"deleted": true,
				"key":     key,
				"value":   value,
			})
		}
	}) //TODO: Require Secrete

	r.Run(":2407")
}

func get(key string, db *bolt.DB) string {
	var value []byte
	err := db.View(func(tx *bolt.Tx) error {
		value = tx.Bucket([]byte("buck")).Get([]byte(key))
		fmt.Printf("Gelesen! Key: %s, Value: %s\n", key, string(value))
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return ""
	}
	return string(value)
}

func put(value string, db *bolt.DB) string {
	var key uint64
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("buck"))
		key, err := bucket.NextSequence()
		if err != nil {
			return fmt.Errorf("Fehler beim key generieren")
		}
		err = bucket.Put([]byte(strconv.FormatUint(key, 10)), []byte(value))
		if err != nil {
			return fmt.Errorf("Fehler beim speichern")
		}
		fmt.Printf("Gespeichert! Key: %d, Value: %s\n", key, value)
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return ""
	}
	return strconv.FormatUint(key, 10)
} //TODO: Shorten function

func remove(key string, db *bolt.DB) {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("buck")).Delete([]byte(key))
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}

/*TODO: write keygen with:
- ultra short
- short
- length
- ultra length
*/
