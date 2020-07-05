package main

import (
        "module/tacplus"
	configure "module/config"
	"log"
	"fmt"
	"strconv"
	"context"
	"strings"
)

type tacasUserCachedNode map[string]interface{}
type tacasUserCached = map[string]tacasUserCachedNode //map[string]interface{}

type tacasHandler struct{
	Config *configure.Config
	UserCached tacasUserCached
}

func (t tacasHandler) HandleAuthenStart(ctx context.Context, a *tacplus.AuthenStart, s *tacplus.ServerSession) *tacplus.AuthenReply {
  if t.Config.LOG.DEBUG.ENABLE==true {
     log.Println("HandleAuthenStart\n")
  }

  log.Println(s.RemoteAddr(),s.LocalAddr(),t.Config.PID)
  user := a.User
  for user == "" {
    if t.Config.LOG.DEBUG.ENABLE==true {
        log.Println("Promting for user...\n")
    }
    c, err := s.GetUser(context.Background(), t.Config.Banner.LoginBanner)
    if err != nil || c.Abort {
      log.Println("Promting for username... %v %v\n",s.RemoteAddr(),s.LocalAddr())
      return nil
    }
    user = c.Message
  }

  pass := ""
  for pass == "" {
    log.Println("Promting for password... ",user,s.RemoteAddr(),s.LocalAddr())
    c, err := s.GetPass(context.Background(), t.Config.Banner.PasswordBanner);
    if err != nil || c.Abort {
      log.Println("Err during GetPass: %v\n", err)
      return nil
    }
    pass = c.Message
  }

  log.Printf("Got: %s:<REMOVED> %v %v \n", user,s.RemoteAddr(),s.LocalAddr())

  if access:=auth(user, pass,t.Config);access!=interface{}(nil) {

    if t.Config.LOG.DEBUG.ENABLE==true {
      log.Println("Replying... ACCESS PERMIT", access)
    }

    switch v := access.(type) {
    default:
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("unexpected type %T %#v\n", v, v)
        }
        log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
        log.Println("INTERNAL ERROR NUMBER 101")
        return &tacplus.AuthenReply{Status: tacplus.AuthenStatusFail, ServerMsg: "INTERNAL ERROR NUMBER 101"};
    case nil:
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("interface{}(nil) type %T\n", v)
        }
        log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
        log.Println("INTERNAL ERROR NUMBER 102")
        return &tacplus.AuthenReply{Status: tacplus.AuthenStatusFail, ServerMsg: "INTERNAL ERROR NUMBER 102"};
    case configure.UserLDAP:
        acs := access.(configure.UserLDAP);
        if Verify_Ip_Access(acs.IPAccess, a.RemAddr )==false {
          log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
          log.Println("INTERNAL ERROR NUMBER 110")
          return &tacplus.AuthenReply{Status: tacplus.AuthenStatusFail, ServerMsg: "ACCESS DENY. POINT 110"};
        }

        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("main.Access type %T\n", v)
          log.Println("Replying... ACCESS PERMIT", acs.USERInfo)
          log.Println("Replying... ACCESS PERMIT", acs.PrivLvl)
          log.Println("Replying... ACCESS PERMIT", acs.Description)
          log.Println("Replying... ACCESS PERMIT", acs.IPAccess)
          log.Println("Replying... ACCESS PERMIT", acs.PERMIT)
        }

        if _, ok := t.UserCached["127.0.0.1"]; ok {
             t.UserCached["127.0.0.1"][user] = acs  //s.LocalAddr()
        }else {
             t.UserCached["127.0.0.1"] = tacasUserCachedNode{}
             t.UserCached["127.0.0.1"][user] = acs
        }
    case configure.User:
        acs := access.(configure.User)
        if Verify_Ip_Access(acs.IPAccess, a.RemAddr )==false {
          log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
          log.Println("INTERNAL ERROR NUMBER 111")
          return &tacplus.AuthenReply{Status: tacplus.AuthenStatusFail, ServerMsg: "ACCESS DENY. POINT 111"};
        }

        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("main.Access type %T\n", v)
          log.Println("Replying... ACCESS PERMIT", acs.PrivLvl)
          log.Println("Replying... ACCESS PERMIT", acs.Description)
          log.Println("Replying... ACCESS PERMIT", acs.IPAccess)
          log.Println("Replying... ACCESS PERMIT", acs.PERMIT)
        }

        if _, ok := t.UserCached["127.0.0.1"]; ok {
             t.UserCached["127.0.0.1"][user] = acs  //s.LocalAddr()
        }else {
             t.UserCached["127.0.0.1"] = tacasUserCachedNode{}
             t.UserCached["127.0.0.1"][user] = acs
        }
     }
     log.Println("Replying... ACCESS PERMIT status:", tacplus.AcctStatusSuccess)
     return &tacplus.AuthenReply{Status: tacplus.AcctStatusSuccess, ServerMsg: t.Config.Banner.BannerAccept };
  } else {
     log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
     return &tacplus.AuthenReply{Status: tacplus.AuthenStatusFail, ServerMsg: t.Config.Banner.BannerReject};
  }
  return &tacplus.AuthenReply{Status: tacplus.AuthenStatusFail, ServerMsg: t.Config.Banner.BannerReject};
}

func (t tacasHandler) HandleAuthorRequest(ctx context.Context, a *tacplus.AuthorRequest) *tacplus.AuthorResponse {

 if t.Config.LOG.DEBUG.ENABLE==true {
     log.Println("Author Request:",a.User, a.Port, a.RemAddr, strings.Join( a.Arg, ", "))
 }
 if acs_ip, ok := t.UserCached["127.0.0.1"]; ok {
  if acs, ok := acs_ip[a.User]; ok {
    privlvl :=""
    PERMIT := []string{}
    if config.LOG.DEBUG.ENABLE==true {
        log.Println("ATH SEARCH in CHADES: ", t.UserCached["127.0.0.1"][a.User] );
        for k, v := range  a.Arg {
            log.Println(k,v)
        }
    }

    arg_parsing := Paring_Args_Request( a.Arg)
    if config.LOG.DEBUG.ENABLE==true {
        log.Printf("Paring_Args_Request() %T %v\n",arg_parsing,arg_parsing)
    }

    switch v := acs.(type) {
    default:
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("unexpected type %T %#v\n", v, v)
        }
        log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
        log.Println("INTERNAL ERROR NUMBER 105")
        return &tacplus.AuthorResponse{Status: tacplus.AuthenStatusFail, ServerMsg: "INTERNAL ERROR NUMBER 105"};
    case nil:
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("interface{}(nil) type %T\n", v)
        }
        log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
        log.Println("INTERNAL ERROR NUMBER 106")
        return &tacplus.AuthorResponse{Status: tacplus.AuthenStatusFail, ServerMsg: "INTERNAL ERROR NUMBER 106"};
    case configure.UserLDAP:
        user := acs.(configure.UserLDAP);
        privlvl=fmt.Sprintf("priv-lvl=%v",strconv.Itoa(user.PrivLvl))
        PERMIT=user.PERMIT
        if Verify_Ip_Access(user.IPAccess, a.RemAddr )==false {
          log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
          log.Println("INTERNAL ERROR NUMBER 120")
          return &tacplus.AuthorResponse{Status: tacplus.AuthenStatusFail, ServerMsg: "ACCESS DENY. POINT 120"};
        }
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("main.Access type %T\n", v)
          log.Println("Replying... ACCESS PERMIT", user.USERInfo)
          log.Println("Replying... ACCESS PERMIT", user.PrivLvl)
          log.Println("Replying... ACCESS PERMIT", user.Description)
          log.Println("Replying... ACCESS PERMIT", user.IPAccess)
          log.Println("Replying... ACCESS PERMIT", user.PERMIT)
          log.Println("Verify_Ip_Access()", Verify_Ip_Access(user.IPAccess, a.RemAddr ))
        }
    case configure.User:
        user := acs.(configure.User)
        privlvl=fmt.Sprintf("priv-lvl=%v",strconv.Itoa(user.PrivLvl))
        PERMIT=user.PERMIT
        if Verify_Ip_Access(user.IPAccess, a.RemAddr )==false {
          log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
          log.Println("INTERNAL ERROR NUMBER 121")
          return &tacplus.AuthorResponse{Status: tacplus.AuthenStatusFail, ServerMsg: "ACCESS DENY. POINT 121"};
        }

        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("main.Access type %T\n", v)
          log.Println("Replying... ACCESS PERMIT", user.PrivLvl)
          log.Println("Replying... ACCESS PERMIT", user.Description)
          log.Println("Replying... ACCESS PERMIT", user.IPAccess)
          log.Println("Replying... ACCESS PERMIT", user.PERMIT)
          log.Println("Verify_Ip_Access()", Verify_Ip_Access(user.IPAccess, a.RemAddr ))
        }
     }


    if len(a.Arg)==2 {
       if arg_parsing["service"]=="shell" && a.Arg[1]=="cmd*" {
         return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusPassAdd, ServerMsg: t.Config.Banner.Banner, Arg: []string{privlvl} };
       } else {
         return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusPassAdd };
       }
    } else if arg_parsing["service"]=="shell" && Verify_Cmd(PERMIT, arg_parsing["cmd"]) {
         return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusPassAdd };
    } else {
         return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail};
    }
    if t.Config.LOG.DEBUG.ENABLE==true {
        log.Println("VALIDATOR COMMAND:",arg_parsing,PERMIT,Verify_Cmd(PERMIT, arg_parsing["cmd"]))
    }
  } else {
      log.Println("USER NOT FOUND in CACHED. ERROR POINT 103.")
      log.Println("SEND ACCESS DENY")
      return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail}

  }
  } else {
      log.Println("USER NOT FOUND in CACHED. ERROR POINT 104.")
      log.Println("SEND ACCESS DENY")
      return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail}

  }

  return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusPassAdd }; //, Arg: []string{"priv-lvl=15"} }; //tacplus.AuthorStatusFail}
}

func (t tacasHandler) HandleAcctRequest(ctx context.Context, a *tacplus.AcctRequest) *tacplus.AcctReply {
 if t.Config.LOG.DEBUG.ENABLE==true {
      log.Println("Acct Request")
 }
 if acs_ip, ok := t.UserCached["127.0.0.1"]; ok {
  if acs, ok := acs_ip[a.User]; ok {
    if config.LOG.DEBUG.ENABLE==true {
        log.Println("ATH SEARCH in CHADES: ", t.UserCached["127.0.0.1"][a.User] );
        for k, v := range  a.Arg {
            log.Println(k,v)
        }
    }

    arg_parsing := Paring_Args_Request( a.Arg)
    if config.LOG.DEBUG.ENABLE==true {
        log.Printf("Paring_Args_Request() %T %v\n",arg_parsing,arg_parsing)
    }
    switch v := acs.(type) {
    default:
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("unexpected type %T %#v\n", v, v)
        }
        log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
        log.Println("INTERNAL ERROR NUMBER 107")
        return &tacplus.AcctReply{Status:  tacplus.AcctStatusError}
    case nil:
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("interface{}(nil) type %T\n", v)
        }
        log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
        log.Println("INTERNAL ERROR NUMBER 108")
        return &tacplus.AcctReply{Status:  tacplus.AcctStatusError}
    case configure.UserLDAP:
        user := acs.(configure.UserLDAP);
        if Verify_Ip_Access(user.IPAccess, a.RemAddr )==false {
          log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
          log.Println("INTERNAL ERROR NUMBER 130")
          return &tacplus.AcctReply{Status:  tacplus.AcctStatusError};
        }

        if _ , ok := arg_parsing["service"]; ok {
            if _, ok := arg_parsing["disc-cause"]; ok {
                if arg_parsing["service"]=="shell" && arg_parsing["disc-cause"]=="1" {
                    delete(t.UserCached["127.0.0.1"],a.User);
                }
            }
        }

        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("main.Access type %T\n", v)
          log.Println("Replying... ACCESS PERMIT", user.USERInfo)
          log.Println("Replying... ACCESS PERMIT", user.PrivLvl)
          log.Println("Replying... ACCESS PERMIT", user.Description)
          log.Println("Replying... ACCESS PERMIT", user.IPAccess)
          log.Println("Replying... ACCESS PERMIT", user.PERMIT)
          log.Println("Verify_Ip_Access()", Verify_Ip_Access(user.IPAccess, a.RemAddr ))
        }
    case configure.User:
        user := acs.(configure.User)
        if Verify_Ip_Access(user.IPAccess, a.RemAddr )==false {
          log.Println("Replying... ACCES DENY", tacplus.AuthenStatusFail)
          log.Println("INTERNAL ERROR NUMBER 131")
          return &tacplus.AcctReply{Status:  tacplus.AcctStatusError};
        }
        if t.Config.LOG.DEBUG.ENABLE==true {
          log.Printf("main.Access type %T\n", v)
          log.Println("Replying... ACCESS PERMIT", user.PrivLvl)
          log.Println("Replying... ACCESS PERMIT", user.Description)
          log.Println("Replying... ACCESS PERMIT", user.IPAccess)
          log.Println("Replying... ACCESS PERMIT", user.PERMIT)
          log.Println("Verify_Ip_Access()", Verify_Ip_Access(user.IPAccess, a.RemAddr ))
       }
     }
  } else {
      log.Println("USER NOT FOUND in CACHED. ERROR POINT 109.")
      log.Println("SEND ACCESS DENY")
      return &tacplus.AcctReply{Status:  tacplus.AcctStatusError}
  }
  } else {
      log.Println("USER NOT FOUND in CACHED. ERROR POINT 110.")
      log.Println("SEND ACCESS DENY")
      return &tacplus.AcctReply{Status:  tacplus.AcctStatusError}
  }
  if t.Config.LOG.DEBUG.ENABLE==true {
      log.Println("REQUEST:",a.User, a.Port, a.RemAddr, strings.Join( a.Arg, ", "))
      for k, v := range  a.Arg {
          log.Println(k,v)
      }
  }
  return &tacplus.AcctReply{Status:  tacplus.AcctStatusSuccess}
}

