package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

//go:generate go build $GOFILE

func main() {
	go http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "http://localhost:8080")
	case "darwin":
		cmd = exec.Command("open", "http://localhost:8080")
	}
	if cmd != nil {
		go func() {
			log.Println("[browser] Open the default browser after two seconds...")
			time.Sleep(time.Second * 2)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}()
	}
	select {}
}

const html = `<!DOCTYPE html>
<html>
<head>
    <title>/math/divide</title>
    <script src="http://code.jquery.com/jquery-latest.js"></script>
</head>
<body>
<form id="logic">  
  <p>a: <input type='number' name='a' value='10'></p>
  <p>b: <input type='number' name='b' value='2'></p>
</form>
<input id="divide" type="button" value="a/b=?" />
<hr>
<p><input id="hosts" type="button" value="Get Gateway Hosts" /></p>
<hr>
<p><a href="http://localhost:5000/home" target="_blank">Go to Home Page</a></p>
<p><a href="http://localhost:5000/home2" target="_blank">Test Redirect to Home Page</a></p>
<script type="application/javascript">
    $('#divide').on('click',function(){
        $.ajax({
            url:"http://localhost:5000/math/divide?seq_=seq_abc_123&access_token=sdfghj",
            type:"POST",
            data:JSON.stringify($('#logic').serializeObject()),
            contentType:"application/json",
            xhrFields: {
                withCredentials: true
            },
            crossDomain: true,
            success:function(res){
                alert(JSON.stringify(res));
            },
            error:function(err){
                alert(JSON.stringify(err));
            }
        });
    });

    $('#hosts').on('click',function(){
        $.ajax({
            url:"http://localhost:5000/gw/v1/hosts?seq_=seq_abc_456",
            type:"GET",
            xhrFields: {
                withCredentials: true
            },
            crossDomain: true,
            success:function(res){
                alert(JSON.stringify(res));
            },
            error:function(err){
                alert(JSON.stringify(err));
            }
        });
    });

    $.fn.serializeObject = function() {
        var o = {};
        var a = this.serializeArray();
        $.each(a, function() {
            if (o[this.name]) {
                if (!o[this.name].push) {
                    o[this.name] = [ o[this.name] ];
                }
                o[this.name].push(parseInt(this.value) || 0);
            } else {
                o[this.name] = parseInt(this.value) || 0;
            }
        });
        return o;
    };
</script>
</body>
</html>`
