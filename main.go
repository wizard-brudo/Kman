package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func EqualBool(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i, value := range a {
		if value != b[i] {
			return false
		}
	}
	return true
}

func ContainsSubStrings(s string, substrings ...string) []bool {
	var arr []bool
	for i := 0; i < len(substrings); i++ {
		arr = append(arr, strings.Contains(s, substrings[i]))
	}
	return arr
}

func DownloadFile(filepath string, url string) error {

	// Получение данных
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Создание файла
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Запись полученных данных в файл
	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			fmt.Println("Проверка устоновленности нужных инструментов")
			cmd, _ := exec.Command("apt", "list", "--installed").Output()
			instruments := ContainsSubStrings(string(cmd), "flex", "bison")
			if EqualBool(instruments, []bool{true, true}) == false {
				fmt.Println("Устоновка нужных инструментов")
				getInstrumentsCmd := exec.Command("sudo", "apt", "install", "bison", "flex", "-y")
				getInstrumentsCmd.Stdin = os.Stdin
				out, _ := getInstrumentsCmd.Output()
				fmt.Println(string(out))
			}
			splitedUrl := strings.Split(os.Args[2], "/")
			tarFilename := splitedUrl[7]
			fmt.Println("Скачивание архива по заданной ссылке")
			DownloadFile(tarFilename, os.Args[2])
			fmt.Println("Распаковка архива с ядром")
			exec.Command("tar", "-xf", tarFilename).Run()
			fmt.Println("Удаление архива с ядром")
			os.Remove(tarFilename)
			fmt.Println("Создание конфигурации сборки")
			splitedTarFilename := strings.Split(tarFilename, ".")
			folder := strings.Join(splitedTarFilename[0:3], ".")
			exec.Command("sh", "-c", "cd "+folder+" && make oldconfig").Run()
			fmt.Println("Сборка ядра linux")
			exec.Command("sh", "-c", "cd "+folder+" && make -j"+strconv.Itoa(runtime.NumCPU())+" bindeb-pkg").Run()
		} else if os.Args[1] == "remove" {
			cmd, _ := exec.Command("sh", "-c", "apt list --installed | egrep 'linux-image|linux-headers'").Output()
			arrayKernels := strings.Split(string(cmd), "\n")
			var kernelToDelete []string
			for _, v := range arrayKernels {
				if strings.Contains(v, os.Args[2]) == true {
					kernelToDelete = append(kernelToDelete, strings.Split(v, "/")[0])
				}
			}
			args := []string{"apt", "purge", "-y"}
			args = append(args, kernelToDelete...)
			fmt.Println("Удаление выбранного ядра")
			DeleteKernelCmd := exec.Command("sudo", args...)
			DeleteKernelCmd.Stdin = os.Stdin
			out, _ := DeleteKernelCmd.Output()
			fmt.Println(string(out))
		} else if os.Args[1] == "list" {
			cmd, _ := exec.Command("sh", "-c", "find /boot/vmli*").Output()
			rawKernelsList := strings.Split(string(cmd), "/boot/vmlinuz")
			KernelsList := rawKernelsList[:len(rawKernelsList)-1]
			fmt.Print("Доступные ядра linux:")
			for _, v := range KernelsList {
				fmt.Print(strings.TrimPrefix(v, "-"))
			}
		} else {
			fmt.Println("Неизвестный аргумент")
		}
	} else {
		fmt.Println("Вы ничего не указали в качестве аргументов")
	}
}
