package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"eardis/storage"
	"eardis/types"
	"eardis/tools"

	jwt "github.com/golang-jwt/jwt/v5"
    "github.com/gorilla/handlers"
)

type APIServer struct{
    address string
    store storage.Storage
}
type ApiError struct {
    Error string
}

type genericHandle func(http.ResponseWriter,*http.Request) error

func genericHandleFunc(f genericHandle) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request){
        err := f(w,r)
        if err != nil{
            tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: err.Error()})
        }
    }
}

func NewAPIServer(address string,store storage.Storage) *APIServer{
    return &APIServer{address: address, store: store}
}

func (s* APIServer) Run() {
    http.ListenAndServe(s.address, handlers.CORS(
        handlers.AllowedOrigins([]string{"http://localhost:5173"}),   // Consentire tutte le origini (da modificare in base alle tue esigenze di sicurezza)
        handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PATCH"}),
        handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With"}), // Aggiungi altri header comuni secondo necessità
        handlers.AllowCredentials(),
    )(s.SetupRoutes()))
}

func (s* APIServer) HandleCheck(w http.ResponseWriter, r *http.Request) error{
    switch r.Method {
        case "GET": return s.check(w,r) 
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}

func (s* APIServer)check(w http.ResponseWriter,r *http.Request) error{
    return tools.WriteJSON(w,http.StatusOK,nil)
}

func (s* APIServer) HandleFiriends(w http.ResponseWriter, r *http.Request,t *jwt.Token) error{
    switch r.Method {
        //case "GET": return s.getUser(w,r,t) 
        //case "POST": return s.searchUser(w,r,t) 
        
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}
func (s* APIServer) HandleNotificationsByID(w http.ResponseWriter, r *http.Request,t *jwt.Token) error{
    switch r.Method {
        case "POST": return s.replyToNotification(w,r,t) 
        
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}

func (s* APIServer) replyToNotification(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    claims := t.Claims.(jwt.MapClaims)
    userid := claims["id"].(string)
    var nresponse types.NotificationResponse
    err := json.NewDecoder(r.Body).Decode(&nresponse);if err != nil {return err}
    defer r.Body.Close()
    switch nresponse.Notification_type{
        case types.FriendRequest: {
            if nresponse.Response{
                err := s.store.AcceptFriendRequest(nresponse.Notification_id,userid); if err!= nil{
                    return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Notification not exist"})
                }else{
                    return tools.WriteJSON(w,http.StatusOK,nil)
                }
            }else{
                err := s.store.DeclineFriendRequest(nresponse.Notification_id,userid); if err!= nil{
                    return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Notification not exist"})
                }else{
                    return tools.WriteJSON(w,http.StatusOK,nil)
                }

            }
            
        }
    }
    return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Invalid Notification type"})
}

func (s* APIServer) HandleNotifications(w http.ResponseWriter, r *http.Request,t *jwt.Token) error{
    switch r.Method {
        case "GET": return s.getNotifications(w,r,t) 
        case "POST": return s.sendNotifications(w,r,t) 
        
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}

func (s* APIServer) getNotifications(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
     claims := t.Claims.(jwt.MapClaims)
     userid := claims["id"].(string)
     notifications,err := s.store.GetNotifications(userid); if err != nil{
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "User not found"})
     }else{
         if len(notifications)<=0{
            return tools.WriteJSON(w,http.StatusOK,nil)
         }else{
            return tools.WriteJSON(w,http.StatusOK,notifications)
         }
     }
}

func (s* APIServer) sendNotifications(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    claims := t.Claims.(jwt.MapClaims)
    userid := claims["id"].(string)
    var message types.Notifications 
    err := json.NewDecoder(r.Body).Decode(&message);if err != nil {return err}
    defer r.Body.Close()
    message.From = userid
    switch message.Type{
        case types.FriendRequest: {
            err := s.store.SendFriendRequestNotifications(message); if err!= nil{
                return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "User does not exist"})
            }else{
                return tools.WriteJSON(w,http.StatusOK,nil)
            }
        }
        default: return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "User does not exist"}) 
        
    }
}

func (s* APIServer) HandleProjects(w http.ResponseWriter, r *http.Request,t *jwt.Token) error{
    switch r.Method {
        //case "GET": return s.getUser(w,r,t) 
        //case "POST": return s.searchUser(w,r,t) 
        
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}

func (s* APIServer) HandleUser(w http.ResponseWriter, r *http.Request,t *jwt.Token) error{
    switch r.Method {
        case "GET": return s.getUser(w,r,t) 
        case "POST": return s.searchUser(w,r,t) 
        case "DELETE": return s.deleteUser(w,r,t)
        
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}

func (s* APIServer) getUser(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
     claims := t.Claims.(jwt.MapClaims)
     userid := claims["id"].(string)
     user,err := s.store.GetUser(userid); if err != nil{
        log.Println(err)
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Account does not exist"})
    }else{
        return tools.WriteJSON(w,http.StatusOK,user)
    }
}

func (s* APIServer) searchUser(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    var search types.SearchUserRequest
    err := json.NewDecoder(r.Body).Decode(&search);if err != nil {return err}
    defer r.Body.Close()
    user,err := s.store.SearchUser(search.Email);if err!= nil{
        log.Println(err)
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Account does not exist"})
    }else{
        return tools.WriteJSON(w,http.StatusOK,user)
    }
}

func (s* APIServer) deleteUser(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    claims := t.Claims.(jwt.MapClaims)
     userid := claims["id"].(string)
     err := s.store.DeleteUser(userid); if err != nil{
        log.Println(err)
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Account does not exist"})
    }else{
         log.Println("[v] User ",userid," deleted")
        return tools.WriteJSON(w,http.StatusOK,nil)
    }
    
}

func (s* APIServer) HandleEvents(w http.ResponseWriter, r *http.Request,t *jwt.Token) error{
    switch r.Method {
        case "GET": return s.getEvents(w,r,t) 
        case "POST": return s.createEvent(w,r,t) 
        //case "PATCH": return s.patchEvent(w,r,t) 
        //case "DELETE": return s.deleteEvent(w,r,t) 
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}

func (s* APIServer) getEvents(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    var events []*types.Event 
    var err error
    claims := t.Claims.(jwt.MapClaims)
    id := claims["id"].(string)
    events, err =  s.store.GetEvents(id); if err!= nil{
        log.Println("Function: getEvents, id: ",id,", Error: ",err)
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "User not found"})
    }else{
        return tools.WriteJSON(w,http.StatusOK,events)
    }
}

func (s* APIServer) createEvent(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    claims := t.Claims.(jwt.MapClaims)
    ownerid := claims["id"].(string)
    var event types.Event = types.Event{Owner: ownerid}
    err := json.NewDecoder(r.Body).Decode(&event);if err != nil {return err}
    defer r.Body.Close()
    s.store.CreateEvent(&event)
    return tools.WriteJSON(w,http.StatusOK,nil)
}


func (s* APIServer) HandleEventById(w http.ResponseWriter, r *http.Request,t *jwt.Token) error{
    switch r.Method {
        case "PATCH": return s.patchEvent(w,r,t) 
        case "DELETE": return s.deleteEvent(w,r,t) 
    }
    return fmt.Errorf("Method not allowed %s", r.Method)
}
func (s* APIServer) patchEvent(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    claims := t.Claims.(jwt.MapClaims)
    ownerid := claims["id"].(string)
    var event types.Event 
    err := json.NewDecoder(r.Body).Decode(&event);if err != nil {return err}
    defer r.Body.Close()
    event.Owner = ownerid
    err = s.store.PatchEvent(ownerid,event.ID,&event); if err != nil{
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "non-existent event, impossible to update"})
    }else{
        return tools.WriteJSON(w,http.StatusOK,nil)
    }
}

func (s* APIServer) deleteEvent(w http.ResponseWriter,r *http.Request,t *jwt.Token) error{
    claims := t.Claims.(jwt.MapClaims)
    ownerid := claims["id"].(string)
    var idRequest types.IdRequest 
    err := json.NewDecoder(r.Body).Decode(&idRequest);if err != nil {return err}
    err =  s.store.DeleteEvent(ownerid,idRequest.ID); if err!= nil{
        log.Println("Function: deleteEvent ","Error: ",err)
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Impossible to delete the event"})
    }else{
        return tools.WriteJSON(w,http.StatusOK,nil)
    }
}

func (s* APIServer) HandleCreateAccount(w http.ResponseWriter, r *http.Request) error{
    if r.Method == "POST"{
        return s.createAccount(w,r) 
    }else{
        return fmt.Errorf("Method not allowed %s", r.Method)
    }
}

func (s* APIServer) createAccount(w http.ResponseWriter,r *http.Request) error{
    var user types.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {return err}
    defer r.Body.Close()
    user.PWD,user.Salt = tools.GeneratePwd(user.PWD) 
    newuser,err := s.store.CreateAccount(&user); if err!=nil{
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "Email or Username already used!"})
    }else{
        cookie := tools.WriteHttpOnlyCookie(newuser.JWT) 
        http.SetCookie(w, &cookie)
        return tools.WriteJSON(w,http.StatusOK,types.TokenResponse{Token: newuser.JWT})
    }
}

func (s* APIServer) HandleLogin(w http.ResponseWriter, r *http.Request) error{
    if r.Method == "POST"{
        return s.login(w,r) 
    }else{
        return fmt.Errorf("Method not allowed %s", r.Method)
    }
}

func (s* APIServer) login(w http.ResponseWriter,r *http.Request) error{
    var user types.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {return err}
    defer r.Body.Close()
    token,err := s.store.Login(&user);if err !=nil{
        return tools.WriteJSON(w,http.StatusBadRequest,ApiError{Error: "User formatting error"})
    }else{
        cookie := tools.WriteHttpOnlyCookie(token) 
        http.SetCookie(w, &cookie)
        return tools.WriteJSON(w,http.StatusOK,types.TokenResponse{Token: token})
    }
}






