import React from "react";
import axios from "axios";

const Home = () => {
  axios({
    method: "POST",
    url: "http://api.localhost:8081/auth/login",
    headers: { Authorization: "{'id': 'test', 'pw': 'test'}" },
  })
    .then((response) => {
      console.log(response);
    })
    .catch((error) => {
      console.log(error);
    });

  return <>HOME</>;
};

export default Home;
