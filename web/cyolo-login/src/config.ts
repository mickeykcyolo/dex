export const AppRoutes = {
    root: "/login",
    mfa: "/mfa",
    enroll: "/enroll",
    redirect: '/redirect'
}

export default {
    ApiVersion: 'v1',
    PollingInteval: 2000,
    AllowAnonymousPageAccess: true,
    Mock : false,
    MockUsername: "asaf",
    MockPassword: "1234",
    MockComputerAlreadySet: false,
    MockPhoneNumber: "1234",
    MockComputerName: "ts-comp",
    MockEnrolled: false,
    MockTotpEnrolled: true,
    MockSupervisor: ""
}