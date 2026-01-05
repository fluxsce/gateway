/**
 * English localization for hub0001 (Login) module
 */
export default {
  login: {
    title: 'Account Login',
    subtitle: 'Please use your credentials to access the system',
    loginTitle: 'Account Login',
    loginSubtitle: 'Please use your credentials to access the system',
    rememberMe: 'Remember me',
    forgotPassword: 'Forgot password?',
    loginButton: 'Login',
    loginSuccess: 'Login successful',
    loginFailed: 'Login failed',
    networkError: 'Network error, please try again later',
    otherLoginMethods: 'Other login methods',
    testAccount: 'Test account: admin / 123456',
    logout: 'Logout',
    logoutSuccess: 'You have been logged out successfully',

    // Login tabs
    accountLogin: 'Account Login',
    phoneLogin: 'Phone Login',

    // User ID and password
    userId: 'User ID',
    username: 'Username',
    usernamePlaceholder: 'Enter your username',
    password: 'Password',
    passwordPlaceholder: 'Enter your password',

    // Captcha related
    captcha: 'Verification Code',
    captchaPlaceholder: 'Enter verification code',

    // Phone verification login
    phoneNumber: 'Phone Number',
    phonePlaceholder: 'Enter your phone number',
    verificationCode: 'Verification Code',
    verificationCodePlaceholder: 'Enter 6-digit code',
    sendCode: 'Send Code',

    // Third-party login
    orLoginWith: 'Or login with',

    // Terms of service
    termsText:
      'By logging in, you agree to our <a href="#">Terms of Service</a> and <a href="#">Privacy Policy</a>',

    codeSent: 'Verification code has been sent',
    codeSendFailed: 'Failed to send verification code',
    wechatRedirect: 'Redirecting to WeChat login...',

    // Welcome content
    welcomeTitle: 'Welcome to Data Management System',
    welcomeSubtitle: 'Efficient, secure, and user-friendly integrated platform',
    featureSecurityTitle: 'Secure & Reliable',
    featureSecurityDesc: 'Advanced encryption technology to ensure your data safety',
    featureAnalyticsTitle: 'Data Analytics',
    featureAnalyticsDesc: 'Powerful analytics tools to help you make informed decisions',
    featureCollaborationTitle: 'Collaboration',
    featureCollaborationDesc: 'Seamless team collaboration features to improve efficiency',

    // WeChat login
    scanQrCode: 'Please scan the QR code to login',
  },
  welcome: {
    title: 'Welcome to Data Management Platform',
    subtitle: 'Efficient and secure data management solution',
    features: {
      secure: 'Secure & Reliable',
      dataAnalysis: 'Data Analysis',
      multiTenant: 'Multi-tenant Support',
    },
    copyright: '© {year} Web Hub - All Rights Reserved',
  },
  forgotPassword: {
    title: 'Reset Password',
    subtitle: 'Enter your email address to receive a password reset link',
    emailPlaceholder: 'Your email address',
    submit: 'Send Reset Link',
    backToLogin: 'Back to Login',
    checkEmail: 'Please check your email for reset instructions',
    emailNotFound: 'Email address not found',
    resetSuccess: 'Your password has been reset successfully',
    resetFailed: 'Failed to reset password',
  },
  validation: {
    userIdRequired: 'User ID is required',
    userIdLength: 'User ID must be between 3-20 characters',
    usernameRequired: 'Username is required',
    usernameLength: 'Username must be between 3-20 characters',
    passwordRequired: 'Password is required',
    passwordLength: 'Password must be between 6-32 characters',
    captchaRequired: 'Verification code is required',
    captchaLength: 'Verification code must be 4 characters',
    phoneRequired: 'Phone number is required',
    phoneFormat: 'Please enter a valid phone number',
    codeRequired: 'Verification code is required',
    codeLength: 'Verification code must be 6 digits',
  },
  captcha: {
    title: 'Verification code',
    refresh: 'Click to refresh',
    placeholder: 'Enter verification code',
  },
  social: {
    wechat: 'Login with WeChat',
    qq: 'Login with QQ',
    github: 'Login with GitHub',
  },
  dashboard: {
    title: 'Dashboard',
    welcome: 'Welcome back, {name}',
    todayStats: "Today's Statistics",
    logins: 'Logins',
    failedAttempts: 'Failed Attempts',
    activeUsers: 'Active Users',
    avgSessionTime: 'Avg. Session Time',
    recentActivity: 'Recent Activity',
    activityTime: 'Time',
    activityAction: 'Action',
    activityIP: 'IP Address',
    activityLocation: 'Location',
    viewAll: 'View All',
  },
  logs: {
    title: 'Login Logs',
    subtitle: 'View all login activity',
    filters: {
      dateRange: 'Date Range',
      searchPlaceholder: 'Search by username, IP, etc.',
      status: 'Status',
      user: 'User',
    },
    table: {
      time: 'Time',
      user: 'User',
      ip: 'IP Address',
      device: 'Device',
      browser: 'Browser',
      status: 'Status',
      location: 'Location',
      actions: 'Actions',
    },
    status: {
      success: 'Success',
      failed: 'Failed',
      locked: 'Locked',
      suspicious: 'Suspicious',
    },
    viewDetails: 'View Details',
    detailsTitle: 'Login Details',
    noData: 'No login records found',
    pagination: {
      prev: 'Previous',
      next: 'Next',
      total: 'Total {total} records',
    },
  },
}
