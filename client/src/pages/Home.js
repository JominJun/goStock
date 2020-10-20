import React, { useState } from "react";
import axios from "axios";

const Home = () => {
  const [loginInfo, setLoginInfo] = useState({
    id: "",
    pw: "",
  });

  const [userInfo, setUserInfo] = useState({
    id: "",
    isAdmin: false,
    money: 0,
    name: "",
    access_token: "",
    isValid: false,
  });

  const getCookieValue = (key) => {
    let cookieKey = key + "=";
    let result = "";
    const cookieArr = document.cookie.split(";");

    for (let i = 0; i < cookieArr.length; i++) {
      if (cookieArr[i][0] === " ") {
        cookieArr[i] = cookieArr[i].substring(1);
      }

      if (cookieArr[i].indexOf(cookieKey) === 0) {
        result = cookieArr[i].slice(cookieKey.length, cookieArr[i].length);
        return result;
      }
    }
    return result;
  };

  const handleLogin = (event) => {
    setLoginInfo({ ...loginInfo, [event.target.name]: event.target.value });
  };

  const enterLogin = (event) => {
    if (event.keyCode === 13) {
      login();
    }
  };

  const login = async () => {
    var params = new URLSearchParams();
    params.append("id", loginInfo.id);
    params.append("pw", loginInfo.pw);

    await axios({
      method: "POST",
      url: "http://api.localhost:8081/v1/auth/login",
      params: params,
    })
      .then((response) => {
        // 쿠키 설정
        var date = new Date();
        date.setDate(date.getDate() + 1);

        document.cookie = `access_token=${
          response.data.access_token
        };Expires=${date.toUTCString()};Secure)`;

        console.log(response.data);

        setUserInfo({
          ...userInfo,
          access_token: response.data.access_token,
        });
      })
      .catch((error) => {
        console.log(error.response.status);
      });
  };

  const logout = () => {
    if (getCookieValue("access_token").length) {
      document.cookie = "access_token=; expires=Thu, 01 Jan 1999 00:00:10 GMT;";
      window.location.reload();
    }
  };

  const checkToken = async () => {
    await axios({
      method: "GET",
      url: "http://api.localhost:8081/v1/auth/validate",
      headers: {
        Authorization: getCookieValue("access_token"),
      },
    })
      .then((response) => {
        setUserInfo({
          ...userInfo,
          id: response.data.result.ID,
          isAdmin: response.data.result.IsAdmin,
          money: response.data.result.Money,
          name: response.data.result.Name,
          isValid: true,
        });
      })
      .catch(() => {
        setUserInfo({
          ...userInfo,
          isValid: false,
        });
      });
  };

  if (getCookieValue("access_token").length) {
    if (userInfo.isValid === false) {
      checkToken();
    }
    if (userInfo.isValid) {
      return (
        <>
          <h1>로그인 성공</h1>
          <button onClick={logout}>로그아웃</button>
        </>
      );
    }
  }

  return (
    <>
      <input
        name="id"
        type="text"
        placeholder="아이디를 입력해주세요"
        onKeyDown={enterLogin}
        onChange={handleLogin}
      />
      <input
        name="pw"
        type="password"
        placeholder="비밀번호를 입력해주세요"
        onKeyDown={enterLogin}
        onChange={handleLogin}
      />
      <button onClick={login}>로그인</button>
    </>
  );
};

export default Home;
