package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/websocket"
)

var (
	addr      = flag.String("addr", "localhost:8080", "http service address")
	upgrader  = websocket.Upgrader{} // use default options
	kafka     *KafkaTopic
	socketMap map[string]Socket
)

type Socket struct {
	Send func(k, v string)
}

func main() {
	flag.Parse()
	kafka = CreateKafkaTopic()
	kafka.Start()
	defer kafka.Close()

	socketMap = make(map[string]Socket)

	// go func() {
	// 	for {
	// 		k, v := kafka.Dequeue()
	// 		if val, ok := socketMap[k]; ok {
	// 			val.Send(k, v)
	// 		}
	// 	}
	// }()

	// go echoJob()
	log.SetFlags(0)

	http.Handle("/:method", func(w http.ResponseWriter, r *http.Request) {
		log.Println("We have something", r.URL.Path)
	})

	// http.HandleFunc("/ws", wsHandler)
	// http.HandleFunc("/", indexHandler)

	log.Fatal(http.ListenAndServe(*addr, nil))
}

//
// func echoJob() {
// 	go func() {
// 		k, v := kafka.Dequeue()
// 		split := strings.Split(k, ":")
// 		if len(split) > 1 {
// 			if split[1] == "echo" {
// 				kafka.Enqueue(split[0]+":response", v)
// 			}
// 		}
// 	}()
// }

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	key := uuid.New()
	log.Printf("WebSocket Connection: %v", c.RemoteAddr())
	socketMap[key+":response"] = Socket{
		Send: func(k, v string) {
			c.WriteMessage(1, []byte(v))
		},
	}

	func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				c.Close()
				delete(socketMap, key+":response")
				break
			}
			log.Printf("message %v\n", string(message))
			kafka.Enqueue(key+":echo", string(message))
		}
	}()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

var indexTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
