import React, { useState, useEffect } from "react";
import { useHistory } from 'react-router-dom'
import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";
import Grid from "@material-ui/core/Grid";
import LockOutlinedIcon from "@material-ui/icons/LockOutlined";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core/styles";
import Alert from "@material-ui/lab/Alert";
import Snackbar from "@material-ui/core/Snackbar"
import CircularProgress from "@material-ui/core/CircularProgress"
import { useUserData } from '../hooks/UserData';
import ScreenContentWrapper from "../components/ScreenContentWrapper";
import { UsernameScheme, PasswordScheme } from "../schemes";
import * as AuthAPI from '../services/AuthAPI'
import Mousetrap from 'mousetrap'
import 'mousetrap-global-bind'
import "./LoginScreen.scss"

const useStyles = makeStyles(theme => ({
  paper: {
    marginTop: theme.spacing(4),
    display: "flex",
    flexDirection: "column",
    alignItems: "center"
  },
  avatar: {
    margin: theme.spacing(1),
    backgroundColor: theme.palette.secondary.main
  },
  form: {
    width: "100%", // Fix IE 11 issue.
    marginTop: theme.spacing(3)
  },
  submit: {
    margin: theme.spacing(3, 0, 2)
  }
}));

const LoginScreen = (): JSX.Element => {
  /* React Hooks to styles and history data */
  const classes = useStyles();

  /* Set state for validation errors */
  const [alertStatus, setAlertStatus] = React.useState({
    show: false,
    severity: "info",
    description: ""
  })
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [validationErrors, setValidationErrors] = useState({
    username: false,
    password: false
  })

  /* Get props and validate them */

  const handleSignIn = async () => {

    const errors = {
      username: Boolean(UsernameScheme.validate(username).error),
      password: Boolean(PasswordScheme.validate(password).error)
    }

    if (Object.values(errors).some(val => val)) {
      //TODO: show error snackbar
      setValidationErrors(errors);
      return;
    }
    setIsLoading(true)
    try {
      const loginResponse = await AuthAPI.performLogin(username, password)
      const { redirect_uri } = loginResponse;
      window.location.href = redirect_uri
    } catch (error) {
      setAlertStatus({
        severity: "error",
        show: true,
        description: error.message
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleUsernameOnChange = (username: string) => {
    setValidationErrors({
      ...validationErrors,
      username: false
    })
    setUsername(username)
  }

  const handlePasswordOnChange = (password: string) => {
    setValidationErrors({
      ...validationErrors,
      password: false
    })
    setPassword(password)
  }

  const handleSnackbarClose = () => {
    setAlertStatus({
      show: false,
      severity: undefined,
      description: ""
    })
  }

  // Keyboard Binds

  useEffect(() => {
    Mousetrap.bindGlobal("enter", () => {
      handleSignIn()
    })
    return () => {
      Mousetrap.unbind("enter")
    }
  })

  const isThereAnyError = Object.values(validationErrors).some(val => val)

  return (
    <div id="root">
      <ScreenContentWrapper>
        <Snackbar open={alertStatus.show} autoHideDuration={6000} onClose={handleSnackbarClose}>
          <Alert onClose={handleSnackbarClose} severity={alertStatus.severity}>
            {alertStatus.description}
          </Alert>
        </Snackbar>
        <div className={classes.paper}>
          <Avatar className={classes.avatar}>
            <LockOutlinedIcon />
          </Avatar>
          <Typography component="h1" variant="h5">
            Login
              </Typography>
          <div className={classes.form}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  variant="outlined"
                  required
                  error={validationErrors.username}
                  fullWidth
                  value={username}
                  onChange={(event) => handleUsernameOnChange(event.target.value)}
                  id="username"
                  label="Username"
                  name="username"

                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  variant="outlined"
                  required
                  fullWidth
                  value={password}
                  error={validationErrors.password}
                  onChange={(event) => handlePasswordOnChange(event.target.value)}
                  name="password"
                  label="Password"
                  type="password"
                  id="password"
                />
              </Grid>
            </Grid>
            {isThereAnyError && <Typography style={{ marginTop: "8px" }} variant="body2" color="error">There is an error with one of the fields</Typography>}
            <Grid container justify = "center" spacing={2}>
              <Grid item xs={6}>
              {
              isLoading
                ? <CircularProgress style={{marginTop: "12px"}} />
                : <Button
                  id="login-submit-btn"
                  type="submit"
                  fullWidth
                  variant="contained"
                  color="primary"
                  className={classes.submit}
                  onClick={handleSignIn}
                >
                  Login
            </Button>
            }
              </Grid>
            </Grid>
          </div>
        </div>
      </ScreenContentWrapper>
    </div>
  );
};

export default LoginScreen;
