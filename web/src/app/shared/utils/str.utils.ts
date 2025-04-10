export class StrUtils {
  static stringifyHTTPErr(err: object): string {
    if (!err) {
      return "Unknown error!!!"
    }

    if (err.hasOwnProperty('message')) {
      // @ts-ignore
      return err['message']
    }

    if (err.hasOwnProperty("error")) {
      // @ts-ignore
      err = err.error
    }
    if (!err.hasOwnProperty("errorMessages")) {
      return "Unknown error!!!"

    }


    let msgs = []
    if (err.hasOwnProperty('httpStatusCode')) {
      // @ts-ignore
      msgs.push(`HTTP status: ${err['httpStatusCode']}. `)
    }

    if (err.hasOwnProperty('errorMessages')) {
      // @ts-ignore
      let errorMessages = err['errorMessages']
      // @ts-ignore
      msgs.push(`Message: ${errorMessages.map(v => v['message']).join(" ")}`)
    }
    return msgs.join("")
  }
}
