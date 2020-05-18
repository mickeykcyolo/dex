import React from 'react';
import { useWindowDimensions } from '../hooks/WindowDimension';
import ReactCodeInput from 'react-verification-code-input';
import './InputCodeField.scss'

interface InputCodeFieldProps {
    fieldNumber?: number,
    onComplete?: (code: string) => void
}

const InputCodeField = ({ fieldNumber=6, onComplete=undefined} : InputCodeFieldProps) => {
    const { height, width } = useWindowDimensions();
    console.log(`width is ${width}`)

    let fieldWidth;
    if(width < 500){
        fieldWidth = width*0.12
    } else if(500 < width && width < 1200){
        fieldWidth = width*0.05
    } else {
        fieldWidth = width*0.03
    }

    return <ReactCodeInput
    fields={fieldNumber}
    onComplete={onComplete}
    className="code-input"
    fieldWidth={fieldWidth} />
}

export default InputCodeField