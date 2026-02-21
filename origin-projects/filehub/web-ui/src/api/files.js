import { api } from './client'

export const listFiles = (params) => api.get('/api/v1/files', { params })

export const getFile = (id) => api.get(`/api/v1/files/${id}`)

export const uploadFile = (file, folderId = null, onProgress) => {
  const form = new FormData()
  form.append('file', file)
  const url = folderId ? `/api/v1/files?folder_id=${folderId}` : '/api/v1/files'
  return api.post(url, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress,
    timeout: 0,
  })
}

export const deleteFile = (id) => api.delete(`/api/v1/files/${id}`)

export const shareFile = (id) => api.get(`/api/v1/files/${id}/share`)

export const getPreviewUrl = (id) => api.get(`/api/v1/files/${id}/preview`)

export const downloadFile = (id, onProgress) =>
  api.get(`/api/v1/files/${id}/download`, {
    responseType: 'blob',
    onDownloadProgress: onProgress,
    timeout: 0,
  })
