import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1' })

api.interceptors.request.use(cfg => {
  const token = localStorage.getItem('token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  const charId = localStorage.getItem('charId')
  if (charId) cfg.headers['X-Character-ID'] = charId
  return cfg
})

api.interceptors.response.use(
  res => res.data,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.reload()
    }
    return Promise.reject(err.response?.data || err)
  }
)

export default api
