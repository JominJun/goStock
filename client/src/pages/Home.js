import React from "react";
import axios from "axios";

const Home = () => {
  var params = new URLSearchParams();
  params.append("id", "test");
  params.append("pw", "test");

  axios({
    method: "POST",
    url: "http://api.localhost:8081/v1/auth/login",
    params: params,
  })
    .then((response) => {
      let access_token = response.data.access_token;
      console.log(access_token);
    })
    .catch((error) => {
      console.log(error.response);
    });

  return <>HOME</>;
};

export default Home;
