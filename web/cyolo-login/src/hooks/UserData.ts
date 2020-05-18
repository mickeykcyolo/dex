import { useState, useEffect } from 'react';
import { IUser } from '../interfaces';
import * as AuthAPI from '../services/AuthAPI';
import Config from '../config'

export function useUserData(){
    const [userData, setUserData] = useState<IUser | undefined>(undefined)

    useEffect(() => {
        const intervalId = setInterval(async () => {
            const userDataResponse = await AuthAPI.getUserData()
            setUserData(userDataResponse);
          }, Config.PollingInteval)

          return () => {
              clearInterval(intervalId);
          }
    }, [])

    return userData;
}