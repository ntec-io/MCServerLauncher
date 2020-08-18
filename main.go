package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/table"
)

type Modpack struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	versions []Version
}

type Version struct {
	ID           int       `json:"id"`
	DisplayName  string    `json:"displayName"`
	FileDate     time.Time `json:"fileDate"`
	ReleaseType  int       `json:"releaseType"`
	DownloadURL  string    `json:"downloadUrl"`
	IsAvailable  bool      `json:"isAvailable"`
	MCVersion    string
	ForgeVersion string
}

func GetPacks(name string) ([]Modpack, error) {
	var modpacks []Modpack

	client := &http.Client{}

	reqString := "https://addons-ecs.forgesvc.net/api/v2/addon/search?categoryId=0&gameId=432&pageSize=255&searchFilter=" + name + "&sectionId=4471&sort=0"

	req, _ := http.NewRequest("GET", reqString, nil)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		fmt.Println(err)
		return modpacks, err
	}

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(respBody, &modpacks)

	return modpacks, nil
}

func (m *Modpack) getVersions() error {
	client := &http.Client{}

	reqstring := "https://addons-ecs.forgesvc.net/api/v2/addon/" + strconv.Itoa(m.ID) + "/files"
	req, _ := http.NewRequest("GET", reqstring, nil)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil
	}

	json.Unmarshal(respBody, &m.versions)
	return nil
}

func inputDial(n int, message string, t table.Writer) int {
	i := 0
	for {
		t.Render()

		fmt.Println(message)
		fmt.Print("Selection: ")
		fmt.Scanf("%d", &i)
		if i > 0 && i < n+1 {
			break
		} else {
			fmt.Println("Pick a valid number!")
		}
	}
	return i - 1
}

func main() {
	var in string
	fmt.Println("Enter Modpack Name")
	fmt.Print("INPUT: ")
	fmt.Scanln(&in)
	in = strings.ReplaceAll(in, " ", "+")

	modpacks, _ := GetPacks(in)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Modpack Name", "Modpack ID"})

	for i, m := range modpacks {
		t.AppendRow(table.Row{i + 1, m.Name, m.ID})
	}

	m := modpacks[inputDial(len(modpacks), "Pick a modpack!", t)]

	m.getVersions()

	sort.Slice(m.versions, func(i, j int) bool {
		return m.versions[i].FileDate.After(m.versions[j].FileDate)
	})

	t = table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Version Name", "Version Date", "Version Type"})

	for i, v := range m.versions {

		rt := ""
		switch v.ReleaseType {
		case 1:
			rt = "Stable"
		case 2:
			rt = "Beta"
		case 3:
			rt = "Alpha"
		default:
			rt = "UNKNOWN"
		}
		t.AppendRow(table.Row{i + 1, v.DisplayName, v.FileDate, rt})
	}

	os.Mkdir("mc", os.ModePerm)

	v := m.versions[inputDial(len(m.versions), "Pick a version!", t)]
	err := DownloadFile("./mc/modpack.zip", v.DownloadURL)
	if err != nil {
		panic(err)
	}
	Unzip("mc/modpack.zip", "mc/")
	os.Remove("mc/modpack.zip")

	file, _ := os.Open("mc/manifest.json")
	bytes, _ := ioutil.ReadAll(file)

	var mani Manifest
	json.Unmarshal(bytes, &mani)

	os.Mkdir("mc/mods", os.ModePerm)

	for _, mods := range mani.Files {
		mods.getModInformation()
		mods.download()
	}

	for _, loader := range mani.Minecraft.ModLoaders {
		if strings.HasPrefix(loader.ID, "forge") && loader.Primary == true {
			v.ForgeVersion = strings.SplitAfter(loader.ID, "-")[1]
		} else {
			panic("Forge is not the primary modloader!")
		}
	}
	v.MCVersion = mani.Minecraft.Version

	fmt.Println("Forge Version:", v.ForgeVersion)
	fmt.Println("MC Version:", v.MCVersion)

	fmt.Println("Download Forge Version")
	forgeURL := fmt.Sprintf("https://files.minecraftforge.net/maven/net/minecraftforge/forge/%[1]s-%[2]s/forge-%[1]s-%[2]s-installer.jar", v.MCVersion, v.ForgeVersion)
	DownloadFile("./mc/forge-installer.jar", forgeURL)

	startForgeInstaller()
	createEula()

	fmt.Println("Server is ready!")
}

func (mod *ModFile) getModInformation() error {
	client := &http.Client{}

	reqstring := "https://addons-ecs.forgesvc.net/api/v2/addon/" + strconv.Itoa(mod.ProjectID) + "/file/" + strconv.Itoa(mod.FileID)
	req, _ := http.NewRequest("GET", reqstring, nil)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil
	}

	json.Unmarshal(respBody, &mod.Information)

	fmt.Println("Processed:", mod.Information.FileName)
	return nil
}

func (f *ModFile) download() {
	fmt.Println("Downloading:", f.Information.FileName)

	DownloadFile("mc/mods/"+f.Information.FileName, f.Information.URL)
}

func startForgeInstaller() {
	fmt.Println("Installing Forge Server")

	cmd := "cd mc && java -jar forge-installer.jar --installServer"
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))
	os.Remove("mc/forge-installer.jar")
}

func createEula() {
	f, _ := os.Create("mc/eula.txt")
	fmt.Println("Accepting EULA (https://account.mojang.com/documents/minecraft_eula)")
	f.WriteString("eula=true")
	defer f.Close()
}
