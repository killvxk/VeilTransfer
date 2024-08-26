package transfer

import (
    "fmt"
    "path"
    "strings"
    "os"
    "github.com/t3rm1n4l/go-mega"
    "VeilTransfer/utils"
)

func UploadMega(username, password, localPath, remoteDir string, includePatterns []string) error {
    m := mega.New()

    err := m.Login(username, password)
    if err != nil {
        return fmt.Errorf("\n[-] failed to login to Mega: %s", err)
    }

    root := m.FS.GetRoot()
    currentNode := root

    if remoteDir != "" {
        remotePathComponents := strings.Split(remoteDir, "/")
        for _, dir := range remotePathComponents {
            nodes, err := m.FS.PathLookup(currentNode, []string{dir})
            if err != nil || len(nodes) == 0 {
                fmt.Printf("[-] Directory %s does not exist, using the last existing directory.\n", dir)
                break
            } else {
                fmt.Printf("[*] Navigating to directory: %s\n\n", dir)
                currentNode = nodes[0]
            }
        }
    }

    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "")

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
            return nil
        }

        name := path.Base(remoteFilePath)

        fmt.Printf("[*] Attempting to upload file: %s to Mega directory: %s\n", localFilePath, currentNode.GetName())

        progress := make(chan int, 1)

        go func() {
            for p := range progress {
                percentage := float64(p) / float64(fileInfo.Size()) * 100
                fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
            }
        }()

        _, err = m.UploadFile(localFilePath, currentNode, name, &progress)
        if err != nil {
            return fmt.Errorf("[-] failed to upload file to Mega: %s", err)
        }

        return nil
    })
}