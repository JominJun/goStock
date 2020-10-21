import React, { useState, useEffect } from "react";
import axios from "axios";
import * as common from "../common/common";

const Home = () => {
  const access_token = common.getCookieValue("access_token");

  const [userInfo, setUserInfo] = useState({
    isLogin: false,
    isAdmin: false,
    id: "",
    name: "",
    money: 0,
  });

  const [loginInfo, setLoginInfo] = useState({
    id: "",
    pw: "",
  });

  const handleChange = (event) => {
    setLoginInfo({ ...loginInfo, [event.target.name]: event.target.value });
  };

  // 처음 로딩할 때
  useEffect(() => {
    // 로그인 체크
    if (access_token === "") {
      setUserInfo({
        ...userInfo,
        isLogin: false,
      });
    } else {
      setUserInfo({
        ...userInfo,
        isLogin: true,
      });
    }
  }, []);

  // access_token을 필요로 하는 작업을 요청할 때 마다
  useEffect(() => {
    // validation check
    if (access_token && userInfo.isLogin) {
      axios({
        method: "GET",
        url: "http://api.localhost:8081/v1/auth/validate",
        headers: {
          Authorization: access_token,
        },
      })
        .then((response) => {
          // if validates
          if (response.data.status === 200) {
            setUserInfo({
              ...userInfo,
              isLogin: true,
              isAdmin: response.data.result.isAdmin,
              id: response.data.result.ID,
              name: response.data.result.Name,
              money: response.data.result.Money,
            });
          }
        })
        .catch(() => {
          common.removeCookie("access_token");
          setUserInfo({
            ...userInfo,
            isLogin: false,
          });
        });

      setUserInfo({
        ...userInfo,
      });
    }
  }, [userInfo.isLogin]);

  const login = () => {
    axios({
      method: "POST",
      url: "http://api.localhost:8081/v1/auth/login",
      params: {
        id: loginInfo.id,
        pw: loginInfo.pw,
      },
    })
      .then((response) => {
        let date = new Date();
        date.setDate(date.getDate() + 1);

        document.cookie = `access_token=${
          response.data.access_token
        }; Expires=${date.toUTCString()}; Secure)`;

        window.location.reload();
      })
      .catch((error) => {
        console.log(error.response.status);
      });
  };

  if (userInfo.isLogin && common.getCookieValue("access_token")) {
    return (
      <>
        <p>아이디: {userInfo.id}</p>
        <p>이름: {userInfo.name}</p>
        <p>돈: {common.numberWithCommas(userInfo.money)}원</p>
        <p>권한: {userInfo.name ? "관리자" : "일반인"}</p>

        <a href="/buystock">
          <button>주식사러가기</button>
        </a>
      </>
    );
  } else {
    return (
      <>
        <input
          name="id"
          type="text"
          placeholder="아이디를 입력해주세요"
          onChange={handleChange}
        />
        <input
          name="pw"
          type="password"
          placeholder="비밀번호를 입력해주세요"
          onChange={handleChange}
        />
        <button onClick={login}>로그인</button>
      </>
    );
  }
};

/*
<ul>
  {list.map((value, index) => (
    <li key={index}>{value}</li>
  ))}
</ul>
*/

export default Home;
