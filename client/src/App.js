import React from "react";
import { BrowserRouter as Router, Route } from "react-router-dom";
import CompanyInfo from "./routes/CompanyInfo";
import Home from "./routes/Home";

const App = () => {
  return (
    <Router>
      <Route path="/" exact component={Home}></Route>
      <Route path="/company" exact component={CompanyInfo}></Route>
    </Router>
  );
};

export default App;
