package main

type Manifest struct {
	Minecraft Minecraft `json:"minecraft"`
	Files     []ModFile `json:files`
}

type Minecraft struct {
	Version    string      `json:"version"`
	ModLoaders []ModLoader `json:"modLoaders"`
}

type ModLoader struct {
	ID      string `json:"id"`
	Primary bool   `json:"primary"`
}

type ModFile struct {
	ProjectID   int  `json:"projectID"`
	FileID      int  `json:"fileID"`
	Required    bool `json:"required"`
	Information FileInformation
}

type FileInformation struct {
	FileName  string `json:"fileName"`
	Available bool   `json:"isAvailable"`
	URL       string `json:"downloadUrl"`
}
