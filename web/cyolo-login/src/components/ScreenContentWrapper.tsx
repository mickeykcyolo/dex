import React, { FunctionComponent } from 'react'
import CssBaseline from "@material-ui/core/CssBaseline";
import Box from "@material-ui/core/Box";
import Container from "@material-ui/core/Container";
import Paper from "@material-ui/core/Paper";
import { makeStyles } from "@material-ui/core/styles";

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

const ScreenContentWrapper : FunctionComponent<{}> = ({ children }) : JSX.Element => {
    const classes = useStyles()

    return (
        <Container maxWidth="sm">
            <CssBaseline/>
            <img src={require('../../public/logo.png')} id="logo" />
            <Box mt={4}>
                <Paper elevation={2}>
                    <Box mt={4} p={3}>
                        { children }
                    </Box>
                </Paper>
            </Box>
            <div style={{ height: "60px"}}></div>
        </Container>
    )
}

export default ScreenContentWrapper