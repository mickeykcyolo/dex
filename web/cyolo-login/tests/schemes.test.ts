import chai from 'chai';

import { PhoneNumberScheme } from '../src/schemes';

decribe('PhoneNumberScheme', () => {
    it('should fail on an invalid phone number', () => {
        const invalidPhoneNumber = "invalidinvalid";
        const { error } = PhoneNumberScheme.validate(invalidPhoneNumber)
        //assert.equal(error, )
    })
})