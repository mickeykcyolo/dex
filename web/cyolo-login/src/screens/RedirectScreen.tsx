import React, { useEffect } from 'react'
import * as UrlBuilder from '../services/UrlBuilder'
import { useHistory } from 'react-router-dom'
import config from '../config'
import ScreenContentWrapper from '../components/ScreenContentWrapper'
import { Typography } from '@material-ui/core'


const RedirectScreen = () => {
    useEffect(() => {
        window.location.href = UrlBuilder.build('redirect')   
    })

    return (
        <div>
            <ScreenContentWrapper>
                <Typography>Redirecting</Typography>
            </ScreenContentWrapper>
        </div>
    )
}

export default RedirectScreen;