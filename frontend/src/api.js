export default {
  GetTemplates(url,jwt) {
    return fetch(url+'/templates', {
      method: 'GET',
      headers: {
        'Authorization': jwt
      }
    })
    .then(res => res.text())
    .then(body => {
      try {
        return JSON.parse(body);
      } catch {
        throw Error(body)
      }
    })
    .then(data => {
      return data.data.templates
    })
    .catch(err => {
      console.log(err.message)
    })
  }
}