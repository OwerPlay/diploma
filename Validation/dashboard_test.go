package main

import (
 "bytes"
 "encoding/json"
 "mime/multipart"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestDiplomaCSVDetection(t *testing.T) {
 cases := []struct{name, file string; headers []string; want string}{
 {"filename_priority", "systemair.csv", []string{"pin_number"}, "Systemair_GFX_v1"},
 {"ambiguous_columns", "unknown.csv", []string{"device_id","prod_date","boot_ver","fw_ver"}, ""},
 {"normalized_headers", "unknown.csv", []string{" PIN_NUMBER "}, "DHWSWEB_v1"},
 {"factory_signature", "unknown.csv", []string{"device_id"}, "TTEHWEB_factory_v1"},
 }
 for _, c := range cases {t.Run(c.name,func(t *testing.T){if got:=detectTemplate(c.file,c.headers);got!=c.want {t.Fatalf("got %q want %q",got,c.want)}})}
}

func TestDiplomaCSVPreview(t *testing.T) {
 var b bytes.Buffer
 mw:=multipart.NewWriter(&b)
 f,_:=mw.CreateFormFile("file","mixed.csv")
 f.Write([]byte("label_name;device_id\nttehweb_label_v1;101\nttehweb_label_factory_id_v2;101\nttehweb_label_v1;102\n"))
 mw.Close()
 r:=httptest.NewRequest("POST","/api/device-labels/preview",&b)
 r.Header.Set("Content-Type",mw.FormDataContentType())
 w:=httptest.NewRecorder();handleBatchPrintPreview(w,r)
 if w.Code!=200 {t.Fatal(w.Code,w.Body.String())}
 var got struct {Rows []map[string]string `json:"rows"`; Groups []struct{Template string `json:"template"`;Count int `json:"count"`} `json:"groups"`}
 if err:=json.Unmarshal(w.Body.Bytes(),&got);err!=nil{t.Fatal(err)}
 if len(got.Rows)!=3 || got.Rows[1]["label_name"]!="ttehweb_label_factory_id_v2" || got.Rows[2]["device_id"]!="102" {t.Fatal(got.Rows)}
 if len(got.Groups)!=2 || got.Groups[0].Count!=2 || got.Groups[1].Template!="TTEHWEB_factory_v1" {t.Fatal(got.Groups)}
}

func TestDiplomaPrintValidation(t *testing.T) {
 cases:=[]struct{name,body string}{
 {"invalid_json","{"},
 {"missing_product",`{}`},
 {"missing_template",`{"product_code":"TEST"}`},
 {"reversed_range",`{"product_code":"TEST","label_template_id":1,"serial_from":104,"serial_to":101}`},
 {"missing_operator",`{"product_code":"TEST","label_template_id":1,"serial_from":101,"serial_to":104,"printed_by":" "}`},
 }
 for _,c:=range cases{t.Run(c.name,func(t *testing.T){w:=httptest.NewRecorder();handlePrint(w,httptest.NewRequest("POST","/api/print",strings.NewReader(c.body)));if w.Code!=400{t.Fatal(w.Code,w.Body.String())}})}
}

func TestDiplomaDateNormalization(t *testing.T){
 for _,s:=range []string{"2026-09-25","25.09.2026","2026-09-25T00:00:00Z"}{t.Run(s,func(t *testing.T){stored,display:=normalizePackedAt(s);if stored!="2026-09-25 00:00:00"||display!="2026-09-25"{t.Fatal(stored,display)}})}
}
