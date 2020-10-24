import React from "react";
import { connect } from "react-redux";
import * as fn from "./common/function";

const MyInfo = ({ myInfo }) => {
  return (
    <>
      <h1>Welcome to MyInfo</h1>
      <p>
        {myInfo.name}({myInfo.id})
      </p>
      <p>{fn.numberWithCommas(myInfo.money)}원</p>
    </>
  );
};

const mapStateToProps = (state) => {
  return { myInfo: state };
};

export default connect(mapStateToProps)(MyInfo);
