package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var todoPath = ""

const DIRSEP = "~"
const DONEPREFIX = "x "

func main() {
	numbPtr := flag.Int("port", 39095, "port number")

	// path to inbox with default value
	inboxDefault := filepath.Join(os.Getenv("HOME"), "Todo")
	flag.StringVar(&todoPath, "path", inboxDefault, "path for executable scripts")
	flag.Parse()

	// check existence path
	_ = os.MkdirAll(todoPath, os.ModePerm)

	// set router
	router := gin.Default()
	router.GET("/quit", quitServer)
	router.POST("/insert", insertTodo) // for call from external (ie. browser extension)

	router.GET("/list", listTodo)
	router.POST("/search", search)
	router.POST("/clear", clear)
	router.POST("/add/:path", addTask)

	router.GET("/list/task/:task", taskSimpleRender)
	router.GET("/list/task/:task/full", taskFullRender)
	router.POST("/list/task/:task/done", taskDone)
	router.POST("/list/task/:task/prioa", prioa)
	router.POST("/list/task/:task/priob", priob)
	router.POST("/list/task/:task/prioc", prioc)
	router.GET("/list/task/:task/edit", taskEdit)
	router.POST("/move/:task/:dir", moveTask)
	router.DELETE("/list/task/:task/delete", taskDelete)

	// set listen port
	portNumberStr := strconv.Itoa(*numbPtr)
	router.Run(":" + portNumberStr)
}

func quitServer(c *gin.Context) {
	c.String(http.StatusOK, "quiting...")
	go func() {
		fmt.Println("quiting launcher...")
		time.Sleep(time.Millisecond * 50)
		os.Exit(0)
	}()
}

func insertTodo(c *gin.Context) {
	title := c.PostForm("title")
	url := c.PostForm("url")

	fmt.Println("data: " + title + " -> " + url)

	title = normalizeFileName(title)
	err := ioutil.WriteFile(filepath.Join(todoPath, "Inbox", title+".txt"), []byte(url), 0644)
	if err != nil {
		log.Println("Unable write file '" + title + ".txt'")
	}

	c.JSON(200, gin.H{
		"status": "posted",
	})
}

// normalize unvanted character
func normalizeFileName(title string) string {
	r := strings.NewReplacer("<", " ",
		">", " ",
		"|", " ",
		"/", " ",
		".", " ",
		":", " ",
		"\\", " ")
	return r.Replace(title)
}

func listTodo(c *gin.Context) {
	body := getIndexList()

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(body))
}

func search(c *gin.Context) {
	search := c.PostForm("search")
	fmt.Println(search)

	// TODO Lebeda - store param to session
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(listTodoDir(todoPath, search, false, "")))
}

func clear(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(listTodoDir(todoPath, "", true, "")))
}

func getIndexList() string {
	body := `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1">
            <meta name="intercoolerjs:use-actual-http-method" content="true"/>
            <title>Todo</title>

            <script src="https://code.jquery.com/jquery-3.1.1.min.js"></script>
            <script src="https://cdn.jsdelivr.net/npm/jquery.hotkeys@0.1.0/jquery.hotkeys.min.js"></script>
			<script src="http://nekman.github.io/keynavigator/keynavigator-min.js"></script>
            <script src="https://intercoolerreleases-leaddynocom.netdna-ssl.com/intercooler-1.2.1.min.js"></script>
            <script defer src="https://use.fontawesome.com/releases/v5.0.3/js/all.js"></script>

            <!-- link rel="stylesheet" href="https://gitcdn.link/repo/Chalarangelo/mini.css/master/dist/mini-default.min.css" -->
            <link rel="stylesheet" href="https://gitcdn.link/repo/Chalarangelo/mini.css/master/dist/mini-dark.min.css">
            <style>
                .hiddencontrols {
                    visibility: hidden;
                }

                .dropdown {
                    position: relative;
                    display: inline-block;
                }

                .dropdown:hover .dropdown-content {
                    display: block;
                }

                /* Dropdown Content (Hidden by Default) */
                .dropdown-content {
                    display: none;
                    position: absolute;
                    background-color: #f9f9f9;
                    min-width: 160px;
                    box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
                    z-index: 1;
                }

                /* Links inside the dropdown */
                .dropdown-content a {
                    color: black;
                    padding: 2px 6px;
                    text-decoration: none;
                    display: block;
                }

                /* Change color of dropdown links on hover */
                .dropdown-content a:hover {background-color: #a1a1a1}
            </style>
        </head>
        <body style="height: 100vh; display: flex; flex-direction: column;">
        <header style="flex-grow: 1; min-height: 44px; display: flex;">
            <input id="searchinput" type="text" style=" padding: 5px; flex-grow: 999;" name="search" ic-post-to="/search" ic-trigger-on="keyup changed"
                           ic-trigger-delay="500ms" ic-target="main" placeholder="search">
            <button type="button" id="searchclear" onclick='$("#searchinput").val(""); $("#searchinput").keyup();'><i class="fas fa-arrow-left"></i></button>
            <button type="button" id="search_a" onclick='$("#searchinput").val("(A) "); $("#searchinput").keyup();'>(A)</button>
            <button type="button" id="clearrepo" ic-post-to="/clear" ic-target="main"><i class="fas fa-trash-alt"></i></button>
        </header>
        <main style="flex-grow: 999; overflow-y: scroll;">
`
	body += listTodoDir(todoPath, "", false, "")
	body += `</main>
        </body>
		<script>
			// shortcuts
			$(document).bind('keydown', 'Alt+g', function(){ $("#searchinput").focus() });
			$(document).bind('keydown', 'Alt+Shift+g', function(){ $("#searchclear").click() });
			$(document).bind('keydown', 'Alt+Shift+a', function(){ $("#search_a").click() });

			// navigate
			$(document).ready(function() {
			  $('li').keynavigator();
			  // $('li').keynavigator(/* optional settings */);
			});
			$('li').keynavigator();
		</script>
        </html>`
	return body
}

/*
"esc","tab","space","return","backspace","scroll","capslock","numlock","insert","home","del","end","pageup","pagedown",
                    "left","up","right","down",
                    "f1","f2","f3","f4","f5","f6","f7","f8","f9","f10","f11","f12",
                    "1","2","3","4","5","6","7","8","9","0",
                    "a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z",
                    "Ctrl+a","Ctrl+b","Ctrl+c","Ctrl+d","Ctrl+e","Ctrl+f","Ctrl+g","Ctrl+h","Ctrl+i","Ctrl+j","Ctrl+k","Ctrl+l","Ctrl+m",
                    "Ctrl+n","Ctrl+o","Ctrl+p","Ctrl+q","Ctrl+r","Ctrl+s","Ctrl+t","Ctrl+u","Ctrl+v","Ctrl+w","Ctrl+x","Ctrl+y","Ctrl+z",
                    "Shift+a","Shift+b","Shift+c","Shift+d","Shift+e","Shift+f","Shift+g","Shift+h","Shift+i","Shift+j","Shift+k","Shift+l",
                    "Shift+m","Shift+n","Shift+o","Shift+p","Shift+q","Shift+r","Shift+s","Shift+t","Shift+u","Shift+v","Shift+w","Shift+x",
                    "Shift+y","Shift+z",
                    "Alt+a","Alt+b","Alt+c","Alt+d","Alt+e","Alt+f","Alt+g","Alt+h","Alt+i","Alt+j","Alt+k","Alt+l",
                    "Alt+m","Alt+n","Alt+o","Alt+p","Alt+q","Alt+r","Alt+s","Alt+t","Alt+u","Alt+v","Alt+w","Alt+x","Alt+y","Alt+z",
                    "Ctrl+esc","Ctrl+tab","Ctrl+space","Ctrl+return","Ctrl+backspace","Ctrl+scroll","Ctrl+capslock","Ctrl+numlock",
                    "Ctrl+insert","Ctrl+home","Ctrl+del","Ctrl+end","Ctrl+pageup","Ctrl+pagedown","Ctrl+left","Ctrl+up","Ctrl+right",
                    "Ctrl+down",
                    "Ctrl+f1","Ctrl+f2","Ctrl+f3","Ctrl+f4","Ctrl+f5","Ctrl+f6","Ctrl+f7","Ctrl+f8","Ctrl+f9","Ctrl+f10","Ctrl+f11","Ctrl+f12",
                    "Shift+esc","Shift+tab","Shift+space","Shift+return","Shift+backspace","Shift+scroll","Shift+capslock","Shift+numlock",
                    "Shift+insert","Shift+home","Shift+del","Shift+end","Shift+pageup","Shift+pagedown","Shift+left","Shift+up",
                    "Shift+right","Shift+down",
                    "Shift+f1","Shift+f2","Shift+f3","Shift+f4","Shift+f5","Shift+f6","Shift+f7","Shift+f8","Shift+f9","Shift+f10","Shift+f11","Shift+f12",
                    "Alt+esc","Alt+tab","Alt+space","Alt+return","Alt+backspace","Alt+scroll","Alt+capslock","Alt+numlock",
                    "Alt+insert","Alt+home","Alt+del","Alt+end","Alt+pageup","Alt+pagedown","Alt+left","Alt+up","Alt+right","Alt+down",
                    "Alt+f1","Alt+f2","Alt+f3","Alt+f4","Alt+f5","Alt+f6","Alt+f7","Alt+f8","Alt+f9","Alt+f10","Alt+f11","Alt+f12"
*/

func listTodoDir(dirName string, search string, clear bool, clientListId string) string {
	body := `<ul id="` + clientListId + `">`

	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		// remove invisible files and dirs
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if file.IsDir() {
			listId := url.PathEscape(strings.Replace(filepath.Join(dirName, file.Name()), "/", DIRSEP, -1))
			clientId := uuid.Must(uuid.NewV4()).String()
			body += `<li><span style="color: darkseagreen; font-size: small; font-style: italic;" 
                    ic-post-to="/add/` + listId + `" ic-target="#` + clientId + `" ic-replace-target="true"        
                    ic-prompt="New task in ` + file.Name() + `:" ic-include='{"cltgt": "` + clientId + `"}'
                    >` + file.Name() + "</span>\n"
			body += listTodoDir(filepath.Join(dirName, file.Name()), search, clear, clientId)
			body += "</li>\n"
		} else {
			// clear
			if clear && strings.HasPrefix(file.Name(), DONEPREFIX) {
				os.Remove(filepath.Join(dirName, file.Name()))
				continue
			}

			// remove by filter
			normSearch := strings.ToLower(strings.TrimSpace(search))    // TODO Lebeda - remove dicritics
			normName := strings.ToLower(strings.TrimSpace(file.Name())) // TODO Lebeda - remove dicritics
			if normSearch != "" && !strings.Contains(normName, normSearch) {
				continue
			}

			// render item
			body += renderSimpleItem(file.Name(), dirName)
		}
	}

	body += "</ul>\n"

	return body
}

func renderSimpleItem(fileName string, dirName string) string {
	style, taskId, clientId := prepareItem(fileName, dirName)
	renderedItem := `<li id="` + clientId +
		`" <span style="` + style + `" ic-get-from="/list/task/` + taskId + `/full" ic-target="#` + clientId + `" ic-replace-target="true">` +
		fileName + `</span> ` +
		"</li>\n"
	return renderedItem
}

func renderFullItem(fileName string, dirName string) string {
	style, taskId, clientId := prepareItem(fileName, dirName)

	content := ""
	fullFileName := filepath.Join(dirName, fileName)
	fileInfo, e := os.Stat(fullFileName)
	fmt.Println(e)
	if (e == nil) && (fileInfo.Size() > 0) && strings.HasSuffix(fileName, ".txt") {
		bytes, e := ioutil.ReadFile(fullFileName)
		if e == nil {
			output := blackfriday.Run(bytes)
			content = string(output[:])
		}
	}

	renderedItem := `<li id="` + clientId +
		`" <span style="` + style + `" ic-get-from="/list/task/` + taskId + `" ic-target="#` + clientId + `" ic-replace-target="true">` +
		fileName + `</span> ` +
		`<div>
            <button type="button" ic-post-to="/list/task/` + taskId + `/done" ic-target="#` + clientId + `" ic-replace-target="true"><i class="fas fa-check"></i></button>
            <button type="button" ic-get-from="/list/task/` + taskId + `/edit"><i class="fas fa-pencil-alt"></i></button>
            <button type="button" ic-post-to="/list/task/` + taskId + `/rename"><i class="far fa-edit"></i></button>
            <button type="button" ic-post-to="/list/task/` + taskId + `/prioa" ic-target="#` + clientId + `" ic-replace-target="true">(A)</button>
            <button type="button" ic-post-to="/list/task/` + taskId + `/priob" ic-target="#` + clientId + `" ic-replace-target="true">(B)</button>
            <button type="button" ic-post-to="/list/task/` + taskId + `/prioc" ic-target="#` + clientId + `" ic-replace-target="true">(C)</button>
            <button type="button" ic-delete-from="/list/task/` + taskId + `/delete" ic-target="#` + clientId + `"><i class="fas fa-trash-alt"></i></a>
           </div>` + content +
		getMoveDirs(taskId) +
		"</li>\n"
	return renderedItem
}

func prepareItem(fileName string, dirName string) (string, string, string) {
	style := ""
	// set style for item
	if strings.HasPrefix(fileName, DONEPREFIX) {
		style += "font-weight: normal; color: #36454c; text-decoration: line-through;" // silver
	} else if strings.HasPrefix(fileName, "(A) ") {
		style += "font-weight: bold;"
	}
	// add item to result
	taskId := url.PathEscape(strings.Replace(filepath.Join(dirName, fileName), "/", DIRSEP, -1))
	clientId := uuid.Must(uuid.NewV4()).String()
	return style, taskId, clientId
}

func getMoveDirs(taskId string) string {
	result := ""
	for _, element := range GetDirectories() {
		element = strings.TrimPrefix(element, todoPath)
		if strings.TrimSpace(element) == "" {
			continue
		}
		pathEscape := url.PathEscape(strings.Replace(element, "/", DIRSEP, -1))
		result += `<button type="button" ic-post-to="/move/` + taskId + `/` + pathEscape + `" ic-target="main">` + element[1:] + `</button>`
		// element is the element from someSlice for where we are
	}
	return result
}

func getTaskFromUrlPath(c *gin.Context) (string, string) {
	task, err := url.PathUnescape(c.Param("task"))
	if err != nil {
		log.Println("Unable render task '" + task)
	}
	task = strings.Replace(task, DIRSEP, "/", -1)
	dir := filepath.Dir(task)
	name := filepath.Base(task)
	return dir, name
}

func taskSimpleRender(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderSimpleItem(name, dir)))
}

func taskFullRender(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderFullItem(name, dir)))
}

func taskDone(c *gin.Context) {
	dir, oldName := getTaskFromUrlPath(c)

	// make new task name
	var newName string
	if strings.HasPrefix(oldName, DONEPREFIX) {
		newName = strings.TrimPrefix(oldName, DONEPREFIX)
	} else {
		// TODO Lebeda - remove priority
		newName = DONEPREFIX + oldName
	}

	// rename file
	err := os.Rename(filepath.Join(dir, oldName), filepath.Join(dir, newName))
	if err != nil {
		fmt.Println(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderSimpleItem(newName, dir)))
}

func prioa(c *gin.Context) {
	taskPrio(c, "A")
}
func priob(c *gin.Context) {
	taskPrio(c, "B")
}
func prioc(c *gin.Context) {
	taskPrio(c, "C")
}

func taskPrio(c *gin.Context, prio string) {
	dir, oldName := getTaskFromUrlPath(c)

	// make new task name
	var newName string
	donePrefix := "(" + prio + ") "
	if strings.HasPrefix(oldName, donePrefix) {
		newName = strings.TrimPrefix(oldName, donePrefix)
	} else {
		// TODO Lebeda - remove priority
		newName = donePrefix + oldName
	}

	// rename file
	err := os.Rename(filepath.Join(dir, oldName), filepath.Join(dir, newName))
	if err != nil {
		fmt.Println(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderSimpleItem(newName, dir)))
}

func taskEdit(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	cmd := exec.Command("exo-open", filepath.Join(dir, name)) // TODO Lebeda - open cmd as param
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func moveTask(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	// decode target directory
	tgtDir, err := url.PathUnescape(c.Param("dir"))
	if err != nil {
		log.Println("Unable render task '" + tgtDir)
	}
	tgtDir = strings.Replace(tgtDir, DIRSEP, "/", -1)
	tgtDir = filepath.Join(todoPath, tgtDir)

	// do rename
	errRename := os.Rename(filepath.Join(dir, name), filepath.Join(tgtDir, name))
	if errRename != nil {
		log.Fatal(errRename)
	}

	// make refresh index
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(listTodoDir(todoPath, "", true, "")))
}

func addTask(c *gin.Context) {
	path, err := url.PathUnescape(c.Param("path"))
	if err != nil {
		log.Println("Unable add task")
	}
	path = strings.Replace(path, DIRSEP, "/", -1)

	cltgt := c.PostForm("cltgt")
	name := c.PostForm("ic-prompt-value")

	_, err = os.Create(filepath.Join(path, normalizeFileName(name)+".txt"))
	if err != nil {
		log.Fatal(err)
	}

	// make refresh index
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(listTodoDir(path, "", false, cltgt)))
}

func taskDelete(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	err := os.Remove(filepath.Join(dir, name))
	if err != nil {
		log.Fatal(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Header().Add("X-IC-Remove", "true")
}

// Returns the names of the subdirectories (including their paths)
// that match the specified search pattern in the specified directory.
func GetDirectories() []string {
	dirs := make([]string, 0, 144)
	filepath.Walk(todoPath, func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() || strings.HasPrefix(fi.Name(), ".") || strings.Contains(path, "/.") {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	return dirs
}
