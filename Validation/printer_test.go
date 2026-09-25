package main

import (
 "bytes"
 "net/http"
 "net/http/httptest"
 "os"
 "strings"
 "testing"
)

func TestDiplomaTemplateVersions(t *testing.T){
 templateCacheDir=t.TempDir()
 requests:=0
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){requests++;w.Header().Set("Content-Disposition",`attachment; filename="example.lbl"`);w.Write([]byte("test fixture"))}));defer s.Close()
 first,err:=ensureTemplate(s.URL,7,1);if err!=nil{t.Fatal(err)}
 again,err:=ensureTemplate(s.URL,7,1);if err!=nil||again!=first||requests!=1{t.Fatal(again,err,requests)}
 next,err:=ensureTemplate(s.URL,7,2);if err!=nil||next==first||requests!=2{t.Fatal(next,err,requests)}
 if _,err:=os.Stat(first);!os.IsNotExist(err){t.Fatal("stale version remains",err)}
}

func TestDiplomaJobContents(t *testing.T){
 templateCacheDir=t.TempDir();jobsDir=t.TempDir()
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Write([]byte("test fixture"))}));defer s.Close()
 job:=PrintJobMsg{JobID:"diploma",TemplateID:1,TemplateVersion:1,Printer:"TEST",Copies:2,Variables:map[string]string{"serial_from":"101","serial_to":"104","description":"Črpalka š ž"}}
 p,err:=buildJobFile(job,s.URL);if err!=nil{t.Fatal(err)}
 data,err:=os.ReadFile(p);if err!=nil{t.Fatal(err)}
 if !bytes.HasPrefix(data,[]byte{0xef,0xbb,0xbf}){t.Fatal("missing UTF8 BOM")}
 text:=string(data)
 if strings.Count(text,"SESSIONPRINT 1, 0")!=4{t.Fatal(text)}
 last:=-1
 for _,serial:=range []string{"101","102","103","104"}{i:=strings.Index(text,`SET serial="`+serial+`"`);if i<=last{t.Fatal(text)};last=i}
 if strings.Contains(text,"SET serial_from=")||!strings.Contains(text,"Črpalka š ž"){t.Fatal(text)}
}

func TestDiplomaTemplateHTTPError(t *testing.T){
 templateCacheDir=t.TempDir()
 s:=httptest.NewServer(http.NotFoundHandler());defer s.Close()
 if _,err:=ensureTemplate(s.URL,1,1);err==nil{t.Fatal("404 accepted")}
}

func TestDiplomaAPIAddress(t *testing.T){
 if got:=apiBaseFromDashboard("wss://example.invalid/ws/printer");got!="https://example.invalid"{t.Fatal(got)}
}
