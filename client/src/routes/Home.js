import React, { useEffect } from "react";
import { connect } from "react-redux";
import { update } from "../store";
import LoginForm from "./LoginForm";
import * as fn from "./common/function";
import axios from "axios";
import MyInfo from "./MyInfo";

const apiDomain = "http://api.localhost:8081/v1/";
const access_token = fn.getCookieValue("access_token");

const Home = ({ myInfo, updateMyInfo }) => {
  useEffect(() => {
    if (access_token.length && myInfo.needValidation) {
      axios({
        url: apiDomain + "auth/validate",
        method: "GET",
        headers: {
          Authorization: access_token,
        },
      })
        .then((response) => {
          updateMyInfo({
            isLogin: true,
            needValidation: false,
            isAdmin: response.data.result.isAdmin,
            id: response.data.result.id,
            name: response.data.name,
            money: response.data.money,
          });
          console.log("store updated");
        })
        .catch((error) => {
          if (error.response.status === 403) {
            fn.removeCookie("access_token");
          }
        });
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [myInfo.needValidation]);

  if (access_token.length) {
    //MyInfo
    return <MyInfo />;
  } else {
    return <LoginForm />;
  }
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
