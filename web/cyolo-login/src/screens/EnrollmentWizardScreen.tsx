import React, { useEffect } from 'react'
import { useHistory } from 'react-router-dom';
import { useUserData } from '../hooks/UserData'
import Config from '../config'

//TODO: fix this import mass
import Check from '@material-ui/icons/Check'
import Close from '@material-ui/icons/Close'
import { makeStyles } from '@material-ui/core/styles';
import Stepper from '@material-ui/core/Stepper'
import Step from '@material-ui/core/Step'
import StepLabel from '@material-ui/core/StepLabel'
import Button from '@material-ui/core/Button'
import Typography from '@material-ui/core/Typography'
import Divider from '@material-ui/core/Divider'
import TextField from '@material-ui/core/TextField'
import Box from '@material-ui/core/Box'
import Avatar from '@material-ui/core/Avatar'
import InputAdornment from '@material-ui/core/InputAdornment'
import ScreenContentWrapper from '../components/ScreenContentWrapper';
import { isMobile } from 'react-device-detect';
import QRCode from 'qrcode.react';
import CircularPrograss from '@material-ui/core/CircularProgress'
import Phone from '@material-ui/icons/Phone'
import red from '@material-ui/core/colors/red';
import green from '@material-ui/core/colors/green';
import Alert from '../components/Alert';
import Snackbar from '@material-ui/core/Snackbar';

import * as AuthAPI from '../services/AuthAPI';
import InputCodeField from '../components/InputCodeField';
import { PhoneNumberScheme } from '../schemes';
import Mousetrap from 'mousetrap';
import './EnrollmentWizardScreen.scss'



const useStyles = makeStyles(theme => ({
    root: {
        width: '100%',
    },
    button: {
        marginRight: theme.spacing(1),
    },
    instructions: {
        marginTop: theme.spacing(1),
        marginBottom: theme.spacing(1),
    },
    title: {
        marginTop: theme.spacing(1),
        marginBottom: theme.spacing(2),
    },
    avatarValid: {
        width: "40px",
        height: "40px",
        backgroundColor: green[500]
    },
    avatarInvalid: {
        alignSelf: 'center',
        width: "40px",
        height: "40px",
        backgroundColor: red[500]
    }
}));

const getSteps = () => [
    "Computer name setup",
    "MFA Setup",
]

interface StepContent {
    element: JSX.Element
    description: string
}

enum AvatarState {
    NotVisible,
    Error,
    Success
}

interface QrViewProps {
    qrCodeValue: string
}
const QrView = ({ qrCodeValue }: QrViewProps) => {
    const classes = useStyles()

    const handleOnEnrollClick = () => {
        window.location.href = qrCodeValue
    }

    return (
        <div>
            {
                isMobile
                    ? <div>
                        <Box p={2}>
                            <Button color="primary" onClick={handleOnEnrollClick} variant="contained" className={classes.button}>Enroll with Authenticator</Button>
                        </Box>
                    </div>
                    : <div style={{ alignItems: 'center', display: '' }}>
                        <Typography variant="h5">Scan the QR Code</Typography>
                        <div style={{ marginTop: "12px", marginBottom: "12px" }}>
                            {
                                qrCodeValue
                                    ? <QRCode
                                        id="123456"
                                        value={qrCodeValue}
                                        size={290}
                                        level={"H"}
                                    />
                                    : <CircularPrograss />
                            }
                        </div>
                    </div>
            }
        </div>
    )
}

interface PhoneEnrollViewProps {
    phoneNumber: string
    enabled?: boolean
    onPhoneChange: (phoneNumber: string) => void
    onSubmitPhone: (phoneNumber: string) => void
}

const PhoneEnrollView = ({ phoneNumber, onPhoneChange, onSubmitPhone, enabled = true }: PhoneEnrollViewProps) => {
    return (
        <div>
            <TextField
                InputProps={{
                    startAdornment: (
                        <InputAdornment position="start">
                            <Phone />
                        </InputAdornment>
                    )
                }}
                value={phoneNumber}
                onChange={(event) => onPhoneChange(event.target.value)}
                placeholder="+972 XXX XXX XXX"
                variant="outlined"
                required
                fullWidth
                id="phone-number"
                label="Phone Number"
                name="phone-number"
            />
            <Box p={2}>
                <Button
                    id="enroll-phone-submit"
                    disabled={!enabled}
                    color="primary"
                    variant="contained"
                    onClick={() => { onSubmitPhone(phoneNumber) }}
                > Submit </Button>
            </Box>
        </div>
    )
}

enum EnrollmentTypes {
    SMS = "sms",
    QR = "totp"
}

const EnrollmentWizardScreen = (): JSX.Element => {
    const classes = useStyles()
    const history = useHistory()
    const userData = useUserData()

    //TODO: check if user is authenticated already, if not redirect to login
    if (!userData && !Config.AllowAnonymousPageAccess) {
        history.push('/')
    }

    // move to the next screen
    if (userData) {
        const { enrolled } = userData;
        if (enrolled) {
            history.push('redirect')
        }
    }

    let enrollmentStatus = {
        sms: true,
        totp: false
    }

    if (userData) {
        enrollmentStatus.sms = Boolean(userData.phone_number)
        enrollmentStatus.totp = Boolean(userData.totp_enrolled)
    }

    const [alertStatus, setAlertStatus] = React.useState({
        show: false,
        severity: "info",
        description: ""
    })

    const [canFinish, setCanFinish] = React.useState(false)
    const [phoneNumber, setPhoneNumber] = React.useState('');
    const [avatarState, setAvatarState] = React.useState(AvatarState.NotVisible)
    const [qrCodeValue, setQrCodeValue] = React.useState("")
    const [computerName, setComputerName] = React.useState("")
    const [isInPhoneSubmissionTimeout, setIsInPhoneSubmissionTimeout] = React.useState(false)
    const [isLoading, setIsLoading] = React.useState(false);
    const [activeStep, setActiveStep] = React.useState(0);
    const [skipped, setSkipped] = React.useState(new Set());
    const [enrollmentMode, setEnrollmentMode] = React.useState<EnrollmentTypes>(EnrollmentTypes.QR)
    const steps = getSteps();

    useEffect(() => {
        (async () => {
            try {
                const userData = await AuthAPI.getUserData()
                if(userData){
    
                    const { personal_desktop, phone_number } = userData
    
                    // check if the user has a configured phone number
                    if(phoneNumber){ 
                        console.log("phone number is defined")
                        setPhoneNumber(phone_number) 
                    }
    
                    // check if the user has a configured computer name
                    if(personal_desktop){
                        setComputerName(personal_desktop)
                        if(activeStep === 0){ handleSkip() }
                    }
                } else {
                    history.push('/')
                }
            } catch (error) {
                console.log('failed loading user data')
            } finally {
                setCanFinish(true)
            }
        })()
    }, [])

    useEffect(() => {
        (async () => {
            const qrCodeValueFromServer = await AuthAPI.getQrCodeUri()
            setQrCodeValue(qrCodeValueFromServer)
        })()
    }, [])

    const isStepOptional = (step: number) => {
        return step === 0;
    };

    const isStepSkipped = (step: number) => {
        return skipped.has(step);
    };

    const handleEnrollmentModeChange = () => {
        setEnrollmentMode(enrollmentMode === EnrollmentTypes.QR ? EnrollmentTypes.SMS : EnrollmentTypes.QR)
    }

    const handleCodeSubmit = async (code: string) => {
        console.log(`code is ${code}`);

        if (Config.Mock) {
            if (enrollmentMode === EnrollmentTypes.SMS) {
                enrollmentStatus.sms = true
            } else {
                enrollmentStatus.totp = true
            }
            setAvatarState(AvatarState.Success)
            setAlertStatus({
                show: true,
                severity: "success",
                description: "successfully verified code"
            })
            return;
        }

        setIsLoading(true)
        try {
            if (enrollmentMode === EnrollmentTypes.SMS) {
                await AuthAPI.performEnrollWithSMS(code)
            } else {
                await AuthAPI.performEnrollWithTOTP(code)
            }

            //TODO: check if this is valid
            setAvatarState(AvatarState.Success)
        } catch (error) {
            setAvatarState(AvatarState.Error)
            setAlertStatus({
                show: true,
                description: "failed to validate code",
                severity: "error"
            })
        } finally {
            setIsLoading(false)
        }
    }

    const handlePhoneSubmission = async (phoneNumber: string) => {

        const { error } = PhoneNumberScheme.validate(phoneNumber);
        if (error) {
            setAlertStatus({
                show: true,
                severity: "error",
                description: "The phone number field is invalid"
            })
            return;
        }

        // throttle the user phone submission
        setIsInPhoneSubmissionTimeout(true)
        setTimeout(() => {
            setIsInPhoneSubmissionTimeout(false)
        }, 3000)


        try {
            await AuthAPI.initiateSmsMessage(phoneNumber);
            console.log("updated phone number");
    
            setAlertStatus({
                show: true,
                severity: "info",
                description: "An sms has been sent to you"
            })   
        } catch (error) {
            setAlertStatus({
                show: true,
                severity: "error",
                description: error.message
            })   
        }
    }

    const renderMfaTypesContentView = () => {
        if (enrollmentMode === EnrollmentTypes.QR) {
            return enrollmentStatus.totp ? <div /> : <QrView qrCodeValue={qrCodeValue} />
        } else {
            return enrollmentStatus.sms ? <div /> : <PhoneEnrollView onPhoneChange={setPhoneNumber} phoneNumber={phoneNumber} enabled={!isInPhoneSubmissionTimeout} onSubmitPhone={handlePhoneSubmission} />
        }
    }

    const renderMfaChangeModeButtonView = () => {
        if (enrollmentMode === EnrollmentTypes.SMS && enrollmentStatus.totp) {
            return <div></div>
        }

        if (enrollmentMode === EnrollmentTypes.QR && enrollmentStatus.sms) {
            return <div></div>
        }

        return (
            <Box p={1}>
                <Button
                    id="enroll-change-mode"
                    color="primary"
                    variant="contained"
                    onClick={handleEnrollmentModeChange}
                    className={classes.button}> Enroll with {enrollmentMode === EnrollmentTypes.QR ? "Phone Number" : "Authenticator"}  </Button>
            </Box>
        )
    }

    const MfaSetupView = (): JSX.Element => {

        let isEnrolled = false
        if (enrollmentMode === EnrollmentTypes.QR) {
            //QR
            isEnrolled = enrollmentStatus.totp
        } else {
            //SMS
            isEnrolled = enrollmentStatus.sms
        }

        return (
            <Box>
                {
                    avatarState !== AvatarState.NotVisible &&
                    <Box p={1}>
                        <div style={{ display: 'flex', justifyContent: 'center' }}>
                            <Avatar className={avatarState === AvatarState.Success ? classes.avatarValid : classes.avatarInvalid}>
                                {
                                    avatarState === AvatarState.Success
                                        ? <Check style={{ color: "white" }} />
                                        : <Close style={{ color: "white" }} />
                                }
                            </Avatar>
                        </div>
                    </Box>
                }
                {
                    renderMfaTypesContentView()
                }

                {
                    renderMfaChangeModeButtonView()
                }
                {
                    !isEnrolled && <Box p={1}>
                        <Divider />
                        <Typography style={{ marginTop: "12px" }} variant="h6">Confirmation Code</Typography>
                        <InputCodeField onComplete={handleCodeSubmit} />
                    </Box>
                }
            </Box>
        )
    }

    const getStepContent = (step: number): StepContent => {
        switch (step) {
            case 0:
                return {
                    element: (
                        <Box m={2} p={2}>
                            <TextField
                                variant="outlined"
                                required
                                fullWidth
                                id="computer-name"
                                value={computerName}
                                onChange={(event) => setComputerName(event.target.value)}
                                label="Computer Name"
                                name="computer-name"
                            />
                        </Box>
                    ),
                    description: "Please input you personal computer name"
                }
            case 1:
                return {
                    element: isLoading ? <CircularPrograss /> : MfaSetupView(),
                    description: "Follow the steps to setup your multi factor authentication"
                }
            default:
                return {
                    element: <div></div>,
                    description: "unknown"
                }
        }
    }

    const updateComputerName = async () => {
        try {
            await AuthAPI.updateComputerName(computerName);
            setAlertStatus({
                show: true,
                severity: "success",
                description: "Computer name was updated"
            })
        } catch (error) {
            setAlertStatus({
                show: true,
                severity: "error",
                description: "failed to update computer name"
            })
        }
    }

    const askServerToSaveUserToDB = async () => {
        await AuthAPI.saveUserToDB()
    }

    const handleNext = () => {
        let newSkipped = skipped;
        if (isStepSkipped(activeStep)) {
            newSkipped = new Set(newSkipped.values());
            newSkipped.delete(activeStep);
        }

        if (activeStep === 0) {

            if (computerName) {
                // update computer name
                updateComputerName()
            } else {
                // if no computer name was specified skip it
                handleSkip()
                return;
            }
        }

        if (activeStep === 1) {

            if (!Object.values(enrollmentStatus).some(val => val)) {
                // all values are false :(
                console.log("I AM")
                return;
            }

            // update enrollment
            askServerToSaveUserToDB()
        }

        setActiveStep(prevActiveStep => prevActiveStep + 1);
        setSkipped(newSkipped);
    };

    const handleBack = () => {
        setActiveStep(prevActiveStep => prevActiveStep - 1);
    };

    const handleSkip = () => {
        if (!isStepOptional(activeStep)) {
            // You probably want to guard against something like this,
            // it should never occur unless someone's actively trying to break something.
            throw new Error("You can't skip a step that isn't optional.");
        }

        setActiveStep(prevActiveStep => prevActiveStep + 1);
        setSkipped(prevSkipped => {
            const newSkipped = new Set(prevSkipped.values());
            newSkipped.add(activeStep);
            return newSkipped;
        });
    };

    const handleSnackbarClose = () => {
        setAlertStatus({
            show: false,
            severity: undefined,
            description: ""
        })
    }

    // Keyboard Binds

    useEffect(() => {
        Mousetrap.bind("enter", () => {
            handleNext()
        })
        return () => {
            Mousetrap.unbind("enter")
        }
    })

    const isEnrolledInSomething = Object.values(enrollmentStatus).some(val => val)
    const canPressNext = (activeStep === steps.length - 1 && isEnrolledInSomething && canFinish) || (activeStep === 0 && computerName.length != 0)

    return (
        <div className={classes.root}>
            <ScreenContentWrapper>
                <Snackbar open={alertStatus.show} autoHideDuration={6000} onClose={handleSnackbarClose}>
                    <Alert onClose={handleSnackbarClose} severity={alertStatus.severity}>
                        {alertStatus.description}
                    </Alert>
                </Snackbar>
                <Typography variant="h4" className={classes.title}>Enrollment</Typography>
                <Stepper activeStep={activeStep}>
                    {steps.map((label, index) => {
                        const stepProps: { completed?: boolean } = {};
                        const labelProps: { optional?: React.ReactNode } = {};
                        if (isStepOptional(index)) {
                            labelProps.optional = <Typography variant="caption">Optional</Typography>;
                        }
                        if (isStepSkipped(index)) {
                            stepProps.completed = false;
                        }
                        return (
                            <Step key={label} {...stepProps}>
                                <StepLabel {...labelProps}>{label}</StepLabel>
                            </Step>
                        );
                    })}
                </Stepper>
                <div>
                    {activeStep === steps.length ? (
                        <div>
                            <Typography className={classes.instructions}>
                                All steps completed - you're finished and will be redirectd in a moment.
              </Typography>
                        </div>
                    ) : (
                            <div>
                                <div>
                                    {getStepContent(activeStep).element}
                                    <Typography className={classes.instructions}>{getStepContent(activeStep).description}</Typography>
                                </div>

                                <div>
                                    <Button id="enroll-back-btn" disabled={activeStep === 0} onClick={handleBack} className={classes.button}>
                                        Back
                </Button>
                                    {isStepOptional(activeStep) && (
                                        <Button
                                            id="enroll-skip-btn"
                                            variant="contained"
                                            color="primary"
                                            onClick={handleSkip}
                                            className={classes.button}
                                        >
                                            Skip
                                        </Button>
                                    )}
                                    {
                                        canPressNext &&
                                        <Button
                                            id="enroll-next-btn"
                                            variant="contained"
                                            color="primary"
                                            onClick={handleNext}
                                            className={classes.button}
                                        >
                                            {activeStep === steps.length - 1 ? 'Finish' : 'Next'}
                                        </Button>
                                    }
                                </div>
                            </div>
                        )}
                </div>
            </ScreenContentWrapper>
        </div>
    )
}

export default EnrollmentWizardScreen