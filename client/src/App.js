import React, { Component } from "react";
import { BrowserRouter, Route } from "react-router-dom";
import Home from "./pages/Home";
import BuyStock from "./pages/BuyStock";

class App extends Component {
  render() {
    return (
      <BrowserRouter>
        <Route exact path="/" component={Home} />
        <Route exact path="/buystock" component={BuyStock} />
      </BrowserRouter>
    );
  }
}

export default App;
