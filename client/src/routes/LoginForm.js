import React, { useState, useEffect } from "react";
import axios from "axios";
import * as fn from "./common/function";

const apiDomain = "http://api.localhost:8081/v1/";

const LoginForm = () => {
  const [loginInfo, setLoginInfo] = useState({
    id: "",
    pw: "",
    isLogining: false,
  });

  useEffect(() => {
    if (loginInfo.isLogining) {
      console.log("로딩 중...");
    } else {
      console.log("로딩 완료!");
    }
  }, [loginInfo.isLogining]);

  const goLogin = (e) => {
    e.preventDefault();

    setLoginInfo({ ...loginInfo, isLogining: true });

    axios({
      url: apiDomain + "auth/login",
      method: "POST",
      params: {
        id: loginInfo.id,
        pw: loginInfo.pw,
      },
    })
      .then((response) => {
        let date = new Date();

        fn.createCookie(
          "access_token",
          response.data.access_token,
          date.toUTCString(date.setHours(date.getDate()))
        );

        setLoginInfo({ ...loginInfo, isLogining: false });
      })
      .catch((error) => {
        switch (error.response.status) {
          case 400:
            console.log("Should Fill Every Content");
            break;
          case 401:
            console.log("Login Failed");
            break;
          default:
            break;
        }

        setLoginInfo({ ...loginInfo, isLogining: false });
      });
  };

  const onChange = (e) => {
    switch (e.target.name) {
      case "id":
        setLoginInfo({ ...loginInfo, id: e.target.value });
        break;
      case "pw":
        setLoginInfo({ ...loginInfo, pw: e.target.value });
        break;
      default:
        break;
    }
  };

  return (
    <>
      {loginInfo.isLogining}
      <form method="POST" onSubmit={goLogin}>
        <input name="id" type="text" placeholder="ID" onChange={onChange} />
        <input name="pw" type="password" placeholder="PW" onChange={onChange} />
        <button type="submit">Login</button>
      </form>
    </>
  );
};

export default LoginForm;
