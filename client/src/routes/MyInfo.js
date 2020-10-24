import React from "react";
import { connect } from "react-redux";

const MyInfo = ({ myInfo }) => {
  console.log(myInfo);
  return (
    <>
      <h1>Welcome to MyInfo</h1>
      <p></p>
    </>
  );
};

const mapStateToProps = (state) => {
  return { myInfo: state };
};

export default connect(mapStateToProps)(MyInfo);
