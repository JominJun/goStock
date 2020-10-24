import React from "react";
import axios from "axios";
import * as fn from "./common/function";

const LoginForm = () => {
  const apiDomain = "http://api.localhost:8081/v1/";

  const loginInfo = {
    id: "",
    pw: "",
  };

  const goLogin = (e) => {
    e.preventDefault();

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
      })
      .catch((error) => {
        switch (error.response.status) {
          case 400:
            console.log("모든 칸을 채워주세요");
            break;
          case 401:
            console.log("로그인 실패");
            break;
          default:
            break;
        }
      });
  };

  const onChange = (e) => {
    switch (e.target.name) {
      case "id":
        loginInfo.id = e.target.value;
        break;
      case "pw":
        loginInfo.pw = e.target.value;
        break;
      default:
        break;
    }
  };

  return (
    <>
      <form method="POST" onSubmit={goLogin}>
        <input name="id" type="text" placeholder="아이디" onChange={onChange} />
        <input
          name="pw"
          type="password"
          placeholder="비밀번호"
          onChange={onChange}
        />
        <button type="submit">로그인</button>
      </form>
    </>
  );
};

export default LoginForm;
