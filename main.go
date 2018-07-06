package main

import (
  "net/http"
  "log"
  "html/template"
  "os"
  "fmt"
  "strings"
  "encoding/json"
  "io/ioutil"
  "net/url"
  "time"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/calendar/v3"
  "github.com/plivo/plivo-go"
)

type Checkbox struct{
	Name string
	Value string
	Ischecked bool
	Text string
}

type PageVariables struct {
  PageTitle        string
  PageCheckbox []Checkbox
  PageCalCheckbox []Checkbox
  Answer           string
}

type Numverify struct {
	Valid               bool   `json:"valid"`
	Number              string `json:"number"`
	LocalFormat         string `json:"local_format"`
	InternationalFormat string `json:"international_format"`
	CountryPrefix       string `json:"country_prefix"`
	CountryCode         string `json:"country_code"`
	CountryName         string `json:"country_name"`
	Location            string `json:"location"`
	Carrier             string `json:"carrier"`
	LineType            string `json:"line_type"`
}

func main() {
  http.HandleFunc("/", DisplayMeetingMember)
  http.HandleFunc("/submited", UserSelected)
  http.HandleFunc("/Calsubmit",CalendarSelected)
  log.Fatal(http.ListenAndServe(getPort(), nil))
}
func getPort() string {
	p := os.Getenv("PORT")
	if p != "" {
		return ":" + p
	}
	return ":9999"
}



func DisplayMeetingMember(w http.ResponseWriter, r *http.Request){
   Title := "Meeting invite"

   MyCheckBox :=[]Checkbox {
	Checkbox{"team","919844577908",true,"Narayana"},
	Checkbox{"team","5678",true,"Manas"},
	Checkbox{"team","9876",true,"Koushik"},
  }
  
  MyCalendarCheckbox := getGoogleCalenderEvent()

  MyPageVariables := PageVariables{
      PageTitle: Title,
      PageCheckbox : MyCheckBox,
      PageCalCheckbox : MyCalendarCheckbox,
    }
   t, err := template.ParseFiles("home.html")
   if err != nil { 
     log.Print("template parsing error: ", err) 
   }
   err = t.Execute(w, MyPageVariables) 
   if err != nil { 
     log.Print("template executing error: ", err) 
   }
  // time.Parse(time.RFC3339, str)
}

func CalendarSelected( w http.ResponseWriter, r *http.Request){
  r.ParseForm()
  fmt.Println("Cal sumbited")
  calMessage := r.Form["cal"]
  productsSelected := r.Form["team"]
  result :=""
  if len(calMessage) != 0{
    if(len(calMessage) == 1){
      result= SendSMSCalEvent(calMessage[0],productsSelected)
   } else
   {
   for _,values :=range calMessage {
    result= SendSMSCalEvent(values,productsSelected)
   }
 }
 }

 MyPageVariables := PageVariables{
  PageTitle: "Hello",
  Answer : result,
  }
  t, err := template.ParseFiles("home.html")
  if err != nil { 
    log.Print("template parsing error: ", err) 
  }
  err = t.Execute(w, MyPageVariables) 
  if err != nil { 
    log.Print("template executing error: ", err) 
}

}


func SendSMSCalEvent(message string, number []string) string {
  result :=""
  if len(number) != 0{
	   if(len(number) == 1){
        if (validatePhoneNumber(number[0]) !=0){
          sendSms(number[0],message)
          result ="SMS sent successfully"
        }else{
        result ="Not a valid phone number"
        }
	  } else
	  {
    for i ,no := range number{
     if (validatePhoneNumber(no) !=0) {
       fmt.Println("Not a valid phone number", no)
       fmt.Println("Not sent a SMS to this number",no)
      number= append(number[:i], number[i+1:]...)
     }
    }
	  finalNumber :=strings.Join(number, "<")
	  fmt.Println("Final numbers are ",finalNumber)
	  sendSms(finalNumber,message)
	  result ="SMS sent successfully"
	}
  }else
  {
  result ="You have not selected any member to send SMS"
  }
 return result
}


func UserSelected(w http.ResponseWriter, r *http.Request){
  r.ParseForm()
  productsSelected := r.Form["team"]
  message :=r.Form.Get("message")
  result :=SendSMSCalEvent(message,productsSelected)

  MyPageVariables := PageVariables{
    PageTitle: "Hello",
    Answer : result,
    }
    t, err := template.ParseFiles("home.html")
    if err != nil { 
      log.Print("template parsing error: ", err) 
    }
    err = t.Execute(w, MyPageVariables) 
    if err != nil { 
      log.Print("template executing error: ", err) 
	}
}

func sendSms(number string,message string){
  client, err := plivo.NewClient("MANTY2ZJI0ZGEXYJI2ZM", "NTUzZDFiMDk2YjI4MWI3YzQ3NzJjOGJkZjE1N2Ew", &plivo.ClientOptions{})
	if err != nil {
		panic(err)
	}
	response, err := client.Messages.Create(
		plivo.MessageCreateParams{
			Src:  "14153336666",
			Dst:  number,
			Text: message,
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Response: %#v\n", response)
 fmt.Println("Numbers are", number)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
  tokFile := "token.json"
  tok, err := tokenFromFile(tokFile)
  if err != nil {
          tok = getTokenFromWeb(config)
          saveToken(tokFile, tok)
  }
  return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Printf("Go to the following link in your browser then type the "+
          "authorization code: \n%v\n", authURL)

  var authCode string
  if _, err := fmt.Scan(&authCode); err != nil {
          log.Fatalf("Unable to read authorization code: %v", err)
  }

  tok, err := config.Exchange(oauth2.NoContext, authCode)
  if err != nil {
          log.Fatalf("Unable to retrieve token from web: %v", err)
  }
  return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  defer f.Close()
  if err != nil {
          return nil, err
  }
  tok := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(tok)
  return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", path)
  f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  defer f.Close()
  if err != nil {
          log.Fatalf("Unable to cache oauth token: %v", err)
  }
  json.NewEncoder(f).Encode(token)
}

func getGoogleCalenderEvent() []Checkbox{
  b, err := ioutil.ReadFile("client_secret.json")
  if err != nil {
          log.Fatalf("Unable to read client secret file: %v", err)
  }

  // If modifying these scopes, delete your previously saved client_secret.json.
  config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
  if err != nil {
          log.Fatalf("Unable to parse client secret file to config: %v", err)
  }

  srv, err := calendar.New(getClient(config))
  if err != nil {
          log.Fatalf("Unable to retrieve Calendar client: %v", err)
  }
  s :=time.Now()
  count := 10
  then := s.Add(time.Duration(-count) * time.Minute)
  after := then.Format(time.RFC3339)
  events, err := srv.Events.List("primary").ShowDeleted(false).
          SingleEvents(true).TimeMin(after).MaxResults(10).OrderBy("startTime").Do()
  if err != nil {
          log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
  }
  TestCal := make([]Checkbox, 0)
  if len(events.Items) == 0 {
          fmt.Println("No upcoming events found.")
  } else {
          for _, item := range events.Items {
                  date := item.Start.DateTime
                  if date == "" {
                          date = item.Start.Date
                  }

                  TestCal = append(TestCal, Checkbox{
                    Name : "cal",
                    Value : item.Summary + date,
                    Ischecked : false,
                    Text :item.Summary + date,
                })
          }
  }
  return TestCal
}

func validatePhoneNumber(number string) int{
  safePhone := url.QueryEscape(number)
	url := fmt.Sprintf("http://apilayer.net/api/validate?access_key=4a1d99b34fa5b1b9882f25eb51e6a0f9&number=%s", safePhone)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return 0
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return 0
	}
	defer resp.Body.Close()
	var record Numverify
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Println(err)
  }
  fmt.Println(record)
  if (!record.Valid && record.LineType !="mobile"){
    return 0
  }
  return 1
}