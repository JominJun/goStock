import React from "react";
import { connect } from "react-redux";

const CompanyInfo = ({ companyInfo }) => {
  return (
    <>
      <h1>Welcome to CompanyInfo</h1>
    </>
  );
};

const mapStateToProps = (state) => {
  return { companyInfo: state };
};

export default connect(mapStateToProps)(CompanyInfo);
