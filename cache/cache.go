package cache

import (
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
)

type Cache struct {
	cacheDir string
}

func NewCache(cacheDir string) *Cache {

	_, err := os.Stat(cacheDir);
	if os.IsNotExist(err) {
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			panic(fmt.Sprintf("failed to create cache directory %s: %v", cacheDir, err))
		}
	}

	return &Cache{
		cacheDir: cacheDir,
	}
}

func (c *Cache) Set(path string, t FileCacheEntry) error {
	pathHash := hex.EncodeToString([]byte(path))

	cacheFile := c.cacheDir + "/" + string(pathHash[:])

	// open or create the cache file
	file, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open cache file %s: %w", cacheFile, err)
	}
	defer file.Close()

	err = gob.NewEncoder(file).Encode(t)
	if err != nil {
		return fmt.Errorf("failed to encode cache file %s: %w", cacheFile, err)
	}


	return nil
}

func (c *Cache) Get(path string) (FileCacheEntry, bool, error) {
	fmt.Println("Getting cache for:", path)

	pathHash := hex.EncodeToString([]byte(path))

	cacheFile := c.cacheDir + "/" + string(pathHash[:])

	file, err := os.Open(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Cache file does not exist:", cacheFile)
			return FileCacheEntry{}, false, nil
		}
		return FileCacheEntry{}, false, fmt.Errorf("failed to open cache file %s: %w", cacheFile, err)
	}
	defer file.Close()

	var t FileCacheEntry
	err = gob.NewDecoder(file).Decode(&t)
	if err != nil {
		return FileCacheEntry{}, false, fmt.Errorf("failed to decode cache file %s: %w", cacheFile, err)
	}

	if t.TargetPath != path {
		return FileCacheEntry{}, false, fmt.Errorf("cache file path mismatch: %s, expected: %s, got: %s", cacheFile, path, t.TargetPath)
	}

	fmt.Println("Cache hit for:", path)
	return t, true, nil
}

type FileCacheEntry struct {
	TargetPath     string
	HashFile [16]byte
	File     []byte
}

func NewTarget(path string, file []byte) FileCacheEntry {
	return FileCacheEntry{
		TargetPath:     path,
		HashFile: md5.Sum(file),
		File:     file,
	}
}
