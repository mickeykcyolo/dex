import React, { useState } from "react";
import "./App.scss";
import { Switch, BrowserRouter as Router, Route, useHistory } from "react-router-dom";
import { AppRoutes } from "./config";
import { createMuiTheme, ThemeProvider } from '@material-ui/core/styles';
import * as AuthAPI from './services/AuthAPI';

const theme = createMuiTheme({
  palette: {
    primary: {
      main: "#123BE5"
    },
    secondary: {
      main: "#D35539"
    },
    background: {
      default: "#EFF2F6"
    }
  },
});

// Screens
import LoginScreen from "./screens/LoginScreen";
import MultifactorAuthScreen from './screens/MultifactorAuthScreen'
import EnrollmentWizardScreen from "./screens/EnrollmentWizardScreen";
import { IUser } from "./interfaces";
import { useUserData } from "./hooks/UserData";
import RedirectScreen from "./screens/RedirectScreen";

const App = () => {
  return (
    <div className="App">
      <ThemeProvider theme={theme}>
        <Router>
          <Switch>
          <Route exact path="/">
              <LoginScreen />
            </Route>
            <Route exact path={AppRoutes.root}>
              <LoginScreen/>
            </Route>
            <Route exact path={AppRoutes.mfa}>
              <MultifactorAuthScreen  />
            </Route>
            <Route exact path={AppRoutes.enroll}>
              <EnrollmentWizardScreen />
            </Route>
            <Route exect path={AppRoutes.redirect}>
              <RedirectScreen/>
            </Route>
          </Switch>
        </Router>
      </ThemeProvider>
    </div>
  );
};

export default App;
