import React, { useState, useEffect } from "react";
import { connect } from "react-redux";
import { update } from "../store";
import axios from "axios";
import * as fn from "./functions/function";

const apiDomain = "http://api.localhost:8081/v1/";

const CompanyInfo = ({ myInfo, updateMyInfo }) => {
  const [isStoreUpdated, setIsStoreUpdated] = useState(false);
  const access_token = fn.getCookieValue("access_token");

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
    console.log(isStoreUpdated);
    if (isStoreUpdated) {
      axios({
        url: apiDomain + "company",
        method: "GET",
        headers: {
          Authorization: access_token,
        },
      })
        .then((response) => {
          console.log(response.data.result);
        })
        .catch((error) => {
          console.log(error);
        });

      return (
        <>
          <h1>Welcome to CompanyInfo</h1>
          <ul></ul>
        </>
      );
    } else {
      //VALIDATION NEEDED
      return <>VALIDATION NEEDED</>;
    }
  } else {
    window.location.href = "/";
    return <>You don't have auth to access.</>;
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

export default connect(mapStateToProps, mapDispatchToProps)(CompanyInfo);
