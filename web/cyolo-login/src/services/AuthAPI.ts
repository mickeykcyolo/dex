import axios, { AxiosError } from 'axios';
import * as qs from 'qs';
import * as UrlBuilder from './UrlBuilder';
import { IUser } from '../interfaces';
import { UserDataScheme } from '../schemes'
import Axios from 'axios';

enum MfaValidationTypes {
    SMS,
    TOTP
}

interface LoginResponse {
    redirect_uri: string 
    user: IUser
}

//TODO: move to outer config file or the like
enum Endpoint {
    Login = "auth/stage/1",
    MfaLoginConfirmation = "auth/stage/2",
    InitateMfaSmsMessage = "auth/stage/2",
    EnrollWithTotpOrSmsCode = "users/me/verify",
    GetQrCodeUri = "users/me/totp-key-uri",
    InitiateEnrollmentSmsMessageCode = "users/me/initiate-sms",
    ConfirmMfaCode = "auth/confirm",
    SaveUser = "users/me/commit",
    GetUserData = "users/me",
    UpdateComputerName = "users/me"
}
/* User */

export const getUserData = async () : Promise<IUser | undefined> => {
    const endpointURL = UrlBuilder.build(Endpoint.GetUserData)
    const { status, data } = await Axios.get<IUser>(endpointURL, { withCredentials: true })
    if(status === 403){
        return undefined;
    }

    const { error } = UserDataScheme.validate(data)
    if(error){
        console.log(error)
        return undefined;
    }
    return data;
}

export const updateComputerName = async (computerName: string) => {
    const endpointURL = UrlBuilder.build(Endpoint.UpdateComputerName)
    await Axios.put(endpointURL, {
        personal_desktop: computerName,
    }, { withCredentials: true })
}

export const saveUserToDB = async () => {
    const endpointURL = UrlBuilder.build(Endpoint.SaveUser);
    await Axios.post(endpointURL, {}, { withCredentials: true })
}

/* Login */

export const performLogin = async (username: string, password: string) : Promise<LoginResponse> => {
    const endpointURL = UrlBuilder.build(Endpoint.Login)
    console.log("endpoint is " + endpointURL)
    try {
        const { data } = await axios.post<LoginResponse>(endpointURL, qs.stringify({
            username,
            password
        }), { withCredentials: true })

        return data;

    } catch (error) {

        const response = error.response
        if(response){
            const { error="user login failed" } = response.data
            throw Error(error)
        }

        throw Error("user login failed")
    }
}

export const performTotpConfirmation = async (totpCode: string) => {
    const endpointURL = UrlBuilder.build(Endpoint.MfaLoginConfirmation)
    await Axios.post(endpointURL, qs.stringify({ totp: totpCode }) , { withCredentials: true })
}

export const initiateMfaSmsMessage = async () => {
    const endpointURL = UrlBuilder.build(Endpoint.InitateMfaSmsMessage)
    await Axios.post(endpointURL, {}, { withCredentials: true })
}

/* Enroll */

export const initiateSmsMessage = async (phoneNumber: string) => {
    const phoneNumberWithNoSpaces = phoneNumber.replace(' ','')

    const endpointURL = UrlBuilder.build(Endpoint.InitiateEnrollmentSmsMessageCode)
    try {
        await Axios.post(endpointURL, { phone_number: phoneNumberWithNoSpaces }, { withCredentials: true })   
    } catch (error) {
        const response = error.response
        if(response){
            const { error="failed to send sms message" } = response.data
            throw Error(error)
        }

        throw Error("failed to send sms message")
    }
}

export const getQrCodeUri = async () : Promise<string> => {
    const endpointURL = UrlBuilder.build(Endpoint.GetQrCodeUri)
    const { data, status } = await Axios.get(endpointURL, { withCredentials: true });
    const { uri } = data;
    return uri
}

export const performEnrollWithTOTP = async (totpCode: string) => {
    await performMfaEnrollment(totpCode, MfaValidationTypes.TOTP)
}

export const performEnrollWithSMS = async (smsCode: string) => {
    await performMfaEnrollment(smsCode, MfaValidationTypes.SMS)
}

const performMfaEnrollment = async (validationCode: string, type: MfaValidationTypes) => {
    let postData = {};
    switch(type){
        case MfaValidationTypes.SMS:
            postData = { kind: "sms", code: validationCode }
            break;
        case MfaValidationTypes.TOTP:
            postData = { kind: "totp", code: validationCode }
            break;
    }

    const endpointURL = UrlBuilder.build(Endpoint.EnrollWithTotpOrSmsCode)
    await axios.post(endpointURL, postData, { withCredentials: true });
}