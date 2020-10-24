import React from "react";
import { connect } from "react-redux";
import { update } from "../store";
import LoginForm from "./LoginForm";
import * as fn from "./common/function";

const access_token = fn.getCookieValue("access_token");

const goValidate = () => {};

const Home = ({ myInfo, updateMyInfo }) => {
  if (access_token.length) {
    // check validation
    return <></>;
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
