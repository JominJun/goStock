import React, { useEffect, useState } from "react";
import { connect } from "react-redux";
import { update } from "../store";
import axios from "axios";

const Home = ({ myInfo, updateMyInfo }) => {
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

  const access_token = getCookieValue("access_token");
  const [userInfo, setUserInfo] = useState({});

  const login = () => {
    axios({
      method: "POST",
      url: "http://api.localhost:8081/v1/auth/login",
      params: {
        id: "test",
        pw: "test",
      },
    })
      .then((response) => {
        let date = new Date();
        date.setDate(date.getDate() + 1);

        document.cookie = `access_token=${
          response.data.access_token
        }; Expires=${date.toUTCString()}; Secure)`;
      })
      .catch((error) => {
        console.log(error.response.status);
      });
  };

  const validate = () => {
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
            isLogin: true,
            isAdmin: response.data.result.isAdmin,
            id: response.data.result.ID,
            name: response.data.result.Name,
            money: response.data.result.Money,
          });
        }
      })
      .catch(() => {});
  };

  useEffect(() => {
    login();
    validate();
  }, []);

  useEffect(() => {
    updateMyInfo(userInfo);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [validate]);

  console.log(myInfo);

  return <></>;
};

const mapStateToProps = (state) => {
  return { myInfo: state };
};

const mapDispatchToProps = (dispatch) => {
  return {
    updateMyInfo: (text) => dispatch(update(text)),
  };
};

export default connect(mapStateToProps, mapDispatchToProps)(Home);
