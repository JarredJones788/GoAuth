package emailer

import (
	"types"

	"gopkg.in/gomail.v2"
)

//Emailer - emailer struct
type Emailer struct {
	Email       string
	Password    string
	SMTPAddress string
	SMTPPort    int
	Host        string
}

//Init - start email service
func (e Emailer) Init(config *types.Config) *Emailer {
	e.Email = config.Email.Email
	e.Password = config.Email.Password
	e.SMTPAddress = config.Email.SMTPAddress
	e.SMTPPort = config.Email.SMTPPort
	e.Host = config.Host
	return &e
}

//NewDeviceEmail - send the account email a new device code
func (e Emailer) NewDeviceEmail(account *types.Account, device *types.Device) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.Email)
	m.SetHeader("To", account.Email)
	m.SetHeader("Subject", "New Device Activation")
	m.SetBody("text/html", e.getTemplate("Your new device code is: <b>"+device.Code+"</b>", "New Login Device", e.Host))

	d := gomail.NewDialer(e.SMTPAddress, e.SMTPPort, e.Email, e.Password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

//RecoverAccountEmail - send a recovery email to the given account
func (e Emailer) RecoverAccountEmail(recovery *types.Recovery) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.Email)
	m.SetHeader("To", recovery.Email)
	m.SetHeader("Subject", "Password Reset")
	m.SetBody("text/html", e.getTemplate("Username: <b>"+recovery.Email+"</b><br/><br/>To reset your password <a href='"+e.Host+"/getRecovery?id="+recovery.ID+"'>Click Here</a>", "Recover Account", e.Host))

	d := gomail.NewDialer(e.SMTPAddress, e.SMTPPort, e.Email, e.Password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

//ChangeEmail - send a request to change email
func (e Emailer) ChangeEmail(emailRequest *types.EmailChange) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.Email)
	m.SetHeader("To", emailRequest.NewEmail)
	m.SetHeader("Subject", "Change Account Email")
	m.SetBody("text/html", e.getTemplate("To Change your email <a href='"+e.Host+"/changeEmail?id="+emailRequest.ID+"'>Click Here</a>", "Change Account Email", e.Host))

	d := gomail.NewDialer(e.SMTPAddress, e.SMTPPort, e.Email, e.Password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func (e Emailer) getTemplate(body string, title string, domain string) string {
	return `
	<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
        <html xmlns="http://www.w3.org/1999/xhtml">
        <head>
            <title>
            </title>
            <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
            <meta name="viewport" content="width=device-width">
            <link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.7.2/css/all.css" integrity="sha384-fnmOCqbTlWIlj8LyTjo7mOUStjsKC4pOpQbqyi7RrhN7udi9RwhKkMHpvLbHG9Sr" crossorigin="anonymous">
            <style type="text/css">body, html {
            margin: 0px;
            padding: 0px;
            -webkit-font-smoothing: antialiased;
            text-size-adjust: none;
            width: 100% !important;
            }
            table td, table {
            }
            #outlook a {
                padding: 0px;
            }
            .ExternalClass, .ExternalClass p, .ExternalClass span, .ExternalClass font, .ExternalClass td, .ExternalClass div {
                line-height: 100%;
            }
            .ExternalClass {
                width: 100%;
            }
            @media only screen and (max-width: 480px) {
                table, table tr td, table td {
                width: 100% !important;
                }
                img {
                width: inherit;
                }
                .layer_2 {
                max-width: 100% !important;
                }
                .edsocialfollowcontainer table {
                max-width: 25% !important;
                }
                .edsocialfollowcontainer table td {
                padding: 10px !important;
                }
                .edsocialfollowcontainer table {
                max-width: 25% !important;
                }
                .edsocialfollowcontainer table td {
                padding: 10px !important;
                }
            }
            </style>
            <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,400i,600,600i,700,700i &subset=cyrillic,latin-ext" data-name="open_sans" rel="stylesheet" type="text/css">
            <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/spectrum/1.8.0/spectrum.min.css">
        </head>
        <body style="padding:0; margin: 0;background: #e4e6ec">
            <table style="height: 100%; width: 100%; background-color: #e4e6ec;" align="center">
            <tbody>
                <tr>
                <td valign="top" id="dbody" data-version="2.31" style="width: 100%; height: 100%; padding-top: 50px; padding-bottom: 50px; background-color: #e4e6ec;">
                    <!--[if (gte mso 9)|(IE)]><table align="center" style="max-width:600px" width="600" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                    <table class="layer_1" align="center" border="0" cellpadding="0" cellspacing="0" style="max-width: 600px; box-sizing: border-box; width: 100%; margin: 0px auto;">
                    <tbody>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #ffffff; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="width: 100%; display: inline-block; vertical-align: top; max-width: 600px;">
                            <table border="0" cellspacing="0" cellpadding="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="emptycell" style="padding: 20px;">
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #ffffff; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="max-width: 596px; display: inline-block; vertical-align: top; width: 100%;">
                            <table class="edcontent" style="border-collapse: collapse;width:100%" border="0" cellpadding="0" cellspacing="0">
                                <tbody>
                                <tr>
                                    <td class="edimg" valign="top" style="padding: 0px; box-sizing: border-box; text-align: center;">
                                    <img style="border-width: 0px; border-style: none; max-width: 255px; width: 100%;" width="255" alt="Image" src="https://competition.boxingontario.com/img/logo.png">
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #ffffff; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="max-width: 596px; display: inline-block; vertical-align: top; width: 100%;">
                            <table border="0" cellspacing="0" cellpadding="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="emptycell" style="padding: 10px;">
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #000000; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="max-width: 600px; display: inline-block; vertical-align: top; width: 100%;">
                            <table border="0" cellspacing="0" cellpadding="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="emptycell" style="padding: 1px;">
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #ffffff; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="max-width: 596px; display: inline-block; vertical-align: top; width: 100%;">
                            <table border="0" cellspacing="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="edtext" style="padding: 20px; text-align: left; color: #5f5f5f; font-size: 12px; font-family: &quot;Open Sans&quot;, &quot;Helvetica Neue&quot;, Helvetica, Arial, sans-serif; word-break: break-word; direction: ltr; box-sizing: border-box;">
                                    <p class="style1 text-center" style="text-align: center; margin: 0px; padding: 0px; color: #000000; font-size: 32px; font-family: &quot;Open Sans&quot;, &quot;Helvetica Neue&quot;, Helvetica, Arial, sans-serif;">` + title + `
                                    </p>
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #ffffff; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="max-width: 596px; display: inline-block; vertical-align: top; width: 100%;">
                            <table border="0" cellspacing="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="edtext" style="padding: 20px; text-align: left; color: #5f5f5f; font-size: 14px; font-family: &quot;Open Sans&quot;, &quot;Helvetica Neue&quot;, Helvetica, Arial, sans-serif; word-break: break-word; direction: ltr; box-sizing: border-box;">
                                    <p class="text-center" style="text-align: center; margin: 0px; padding: 0px;">` + body + `
                                    </p>
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <div class="layer_2" style="max-width: 596px; display: inline-block; vertical-align: top; width: 100%;">
                            <table border="0" cellspacing="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="edtext" style="padding: 20px; text-align: left; color: #5f5f5f; font-size: 12px; font-family: &quot;Open Sans&quot;, &quot;Helvetica Neue&quot;, Helvetica, Arial, sans-serif; word-break: break-word; direction: ltr; box-sizing: border-box;">
                                    <p class="text-center" style="text-align: center; margin: 0px; padding: 0px;">Having trouble finding your account? Please go <a href="` + domain + `/login">here</a> and click forgot password.
                                    </p>
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #ffffff; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="max-width: 596px; display: inline-block; vertical-align: top; width: 100%;">
                            <table border="0" cellspacing="0" cellpadding="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="emptycell" style="padding: 10px;">
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #f4f4f3; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="display: inline-block; vertical-align: top; width: 100%; max-width: 600px;">
                            <table border="0" cellspacing="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="edsocialfollow" style="padding: 20px;">
                                    <table align="center" style="margin:0 auto" class="edsocialfollowcontainer" cellpadding="0" border="0" cellspacing="0">
                                        <tbody>
                                        <tr>
                                            <td>
                                            <!--[if mso]><table align="center" border="0" cellspacing="0" cellpadding="0"><tr><td align="center" valign="top"><![endif]-->
                                            <table align="left" border="0" cellpadding="0" cellspacing="0" data-service="facebook">
                                                <tbody>
                                                <tr>
                                                    <td align="center" valign="middle" style="padding:10px;">
                                                    <a href="https://www.facebook.com/boxingontario" target="_blank" style="color:#5457ff;font-size:12px;font-family:"><img src="https://api.etrck.com/userfile/a18de9fc-4724-42f2-b203-4992ceddc1de/ro_sol_co_32_facebook.png" style="display:block;width:32px;max-width:32px;border:none" alt="Facebook"></a></td>
                                                </tr>
                                                </tbody>
                                            </table>
                                            <!--[if mso]></td><td align="center" valign="top"><![endif]-->
                                            <table align="left" border="0" cellpadding="0" cellspacing="0" data-service="twitter">
                                                <tbody>
                                                <tr>
                                                    <td align="center" valign="middle" style="padding:10px;">
                                                    <a href="https://twitter.com/BoxingOntario" target="_blank" style="color:#5457ff;font-size:12px;font-family:"><img src="https://api.etrck.com/userfile/a18de9fc-4724-42f2-b203-4992ceddc1de/ro_sol_co_32_twitter.png" style="display:block;width:32px;max-width:32px;border:none" alt="Twitter"></a></td>
                                                </tr>
                                                </tbody>
                                            </table>
                                            <!--[if mso]></td></tr></table><![endif]-->
                                            </td>
                                        </tr>
                                        </tbody>
                                    </table>
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                        <tr>
                        <td class="drow" valign="top" align="center" style="background-color: #ffffff; box-sizing: border-box; font-size: 0px; text-align: center;">
                            <!--[if (gte mso 9)|(IE)]><table width="100%" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td valign="top"><![endif]-->
                            <div class="layer_2" style="max-width: 596px; display: inline-block; vertical-align: top; width: 100%;">
                            <table border="0" cellspacing="0" cellpadding="0" class="edcontent" style="border-collapse: collapse;width:100%">
                                <tbody>
                                <tr>
                                    <td valign="top" class="emptycell" style="padding: 10px;">
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                            </div>
                            <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                        </td>
                        </tr>
                    </tbody>
                    </table>
                    <!--[if (gte mso 9)|(IE)]></td></tr></table><![endif]-->
                </td>
                </tr>
            </tbody>
            </table>
        </body>
        </html>
	`
}
