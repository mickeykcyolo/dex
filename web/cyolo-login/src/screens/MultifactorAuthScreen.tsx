import React, { useState, useEffect } from "react";
import Avatar from "@material-ui/core/Avatar";
import LockOutlinedIcon from "@material-ui/icons/LockOutlined";
import { makeStyles } from "@material-ui/core/styles";
import Alert from "@material-ui/lab/Alert";
import {
  Box,
  Typography,
  Link,
  CircularProgress,
  Snackbar
} from "@material-ui/core";
import { useHistory } from "react-router-dom";
import ScreenContentWrapper from "../components/ScreenContentWrapper";
import config from "../config";
import * as AuthAPI from "../services/AuthAPI";
import { useUserData } from "../hooks/UserData";
import InputCodeField from "../components/InputCodeField";
import "./MultifactorAuthScreen.scss";

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
    marginTop: theme.spacing(3),
    alignItems: "center"
  },
  submit: {
    margin: theme.spacing(3, 0, 2)
  },
  numberInput: {
    backgroundColor: theme.palette.secondary.main
  }
}));

enum MfaModes {
  SMS = "sms",
  TTOP = "ttop"
}

const MultifactorAuthScreen = (): JSX.Element => {
  const userData = useUserData();
  const history = useHistory();
  const classes = useStyles();

  const [isLoading, setIsLoading] = useState(true);
  const [mfaMode, setMfaMode] = useState(MfaModes.TTOP);
  const [alertStatus, setAlertStatus] = React.useState({
    show: false,
    severity: "info",
    description: ""
  });

  if (!userData && !config.AllowAnonymousPageAccess) {
    history.push("/");
  }

  // in charge of redirecting to the next screen
  if (userData) {
    // user is defined
    const { auth_level, name } = userData;

    // check if it has mfa activated
    if (name !== "anonymous" && auth_level > 1) {
      history.push("redirect");
    }
  }

  let validMfaModes: MfaModes[] = [];
  if (userData?.totp_enrolled && !userData?.supervisor) {
    validMfaModes.push(MfaModes.TTOP);
  }
  if (userData?.phone_number || userData?.supervisor) {
    validMfaModes.push(MfaModes.SMS);
  }

  useEffect(() => {
    (async () => {
      const userData = await AuthAPI.getUserData();
      if (userData) {
        const { enrolled } = userData;
        if (!enrolled) {
          history.push("/");
        }
      } else {
        history.push("/");
      }
    })();
  }, []);

  useEffect(() => {
    setTimeout(() => {
      setIsLoading(false);
    }, 2000);
  });

  //TODO: name this function better
  const initiateSmsSending = async () => {
    if (config.Mock) {
      setAlertStatus({
        show: true,
        severity: "info",
        description: "an sms message was sent"
      });
      return;
    }

    try {
      await AuthAPI.initiateMfaSmsMessage();
      setAlertStatus({
        show: true,
        severity: "info",
        description: "an sms message was sent"
      });
    } catch (error) {
      setAlertStatus({
        show: true,
        severity: "error",
        description: "failed to send sms message"
      });

      setTimeout(() => {
        // if an error happened return to the previous mode
        setMfaMode(MfaModes.TTOP);
      }, 1500);
    }
  };

  const handleModeChange = () => {
    // invert the mfa modes
    const newMode = mfaMode === MfaModes.TTOP ? MfaModes.SMS : MfaModes.TTOP;
    setMfaMode(newMode);

    if (newMode === MfaModes.SMS) {
      console.log("sending sms message");
      initiateSmsSending();
    }
  };

  const handleCodeCompletion = async (code: string) => {
    if (config.Mock) {
      setAlertStatus({
        severity: "success",
        show: true,
        description: "code verified successfully"
      });
      return;
    }

    // filter out people with supervisor
    if (userData?.supervisor) {
      setAlertStatus({
        severity: "info",
        show: true,
        description: "users with a supervisor set, must connect with sms"
      });
      return;
    }

    setIsLoading(true);
    try {
      await AuthAPI.performTotpConfirmation(code);
      setAlertStatus({
        severity: "success",
        show: true,
        description: "code verified successfully"
      });
    } catch (error) {
      setAlertStatus({
        severity: "error",
        show: true,
        description: "failed to verify code"
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleSnackbarClose = () => {
    setAlertStatus({
      show: false,
      severity: undefined,
      description: ""
    });
  };

  //TODO: check against user for sms existance and show link accordingly
  const renderTtopMfaView = () => {
    return (
      <div>
        {validMfaModes.includes(MfaModes.TTOP) && (
          <InputCodeField onComplete={handleCodeCompletion} />
        )}
        {validMfaModes.includes(MfaModes.SMS) && (
          <Box m={1}>
            <Link onClick={handleModeChange} href="#">
              Click here for SMS Authentication
            </Link>
          </Box>
        )}
      </div>
    );
  };

  const renderSmsMfaView = () => {
    const { supervisor = "" } = userData || {};

    return (
      <div>
        {supervisor ? (
          <Typography variant="h6">
            An sms has been sent to your supervisor with a link,
            <br />
            Please ask him to follow its instructions and verify your mfa status
          </Typography>
        ) : (
          <Typography variant="h6">
            An sms has been sent to you with a link,
            <br />
            Please follow its instructions to verify your mfa status
          </Typography>
        )}
        <Box p={2}>
          <CircularProgress />
        </Box>
        {validMfaModes.includes(MfaModes.TTOP) && (
          <Link href="#" onClick={handleModeChange}>
            Click here for Token Authentication
          </Link>
        )}
      </div>
    );
  };

  const renderContentMode = (mode: MfaModes) => {
    switch (mode) {
      case MfaModes.SMS:
        return renderSmsMfaView();
      case MfaModes.TTOP:
        return renderTtopMfaView();
    }
  };

  return (
    <div id="root">
      <ScreenContentWrapper>
        <Snackbar
          open={alertStatus.show}
          autoHideDuration={6000}
          onClose={handleSnackbarClose}
        >
          <Alert onClose={handleSnackbarClose} severity={alertStatus.severity}>
            {alertStatus.description}
          </Alert>
        </Snackbar>
        <div className={classes.paper}>
          <Avatar className={classes.avatar}>
            <LockOutlinedIcon />
          </Avatar>
          <Box>
            <Typography component="h1" variant="h5">
              Multifactor Verification
            </Typography>
            {isLoading ? (
              <CircularProgress style={{ marginTop: "12px" }} />
            ) : (
              renderContentMode(mfaMode)
            )}
          </Box>
        </div>
      </ScreenContentWrapper>
    </div>
  );
};

export default MultifactorAuthScreen;
