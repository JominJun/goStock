import React, { useEffect, useState } from "react";
import { connect } from "react-redux";
import { update } from "../store";
import LoginForm from "./LoginForm";
import * as fn from "./common/function";
import axios from "axios";
import MyInfo from "./MyInfo";

const apiDomain = "http://api.localhost:8081/v1/";
const access_token = fn.getCookieValue("access_token");

const Home = ({ myInfo, updateMyInfo }) => {
  const [isStoreUpdated, setIsStoreUpdated] = useState(false);

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
          let res = response.data.result;
          updateMyInfo({
            isLogin: true,
            needValidation: false,
            isAdmin: res.IsAdmin,
            id: res.ID,
            name: res.Name,
            money: res.Money,
          }).then(() => {
            setIsStoreUpdated(true);
          });
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
    if (isStoreUpdated) {
      return <MyInfo />;
    }
    return <>Page Reload Needed</>;
  } else {
    return <LoginForm />;
  }
};

const mapStateToProps = (state) => {
  return { myInfo: state };
};

const mapDispatchToProps = (dispatch) => {
  return {
    updateMyInfo: async (text) => dispatch(update(text)),
  };
};

export default connect(mapStateToProps, mapDispatchToProps)(Home);
