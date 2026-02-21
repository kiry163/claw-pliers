import { api } from './client'

// 创建文件夹
export function createFolder(data) {
  return api.post('/api/v1/folders', data)
}

// 列出文件夹
export function listFolders(parentId = null) {
  const params = parentId ? { parent_id: parentId } : {}
  return api.get('/api/v1/folders', { params })
}

// 获取文件夹内容
export function getFolderContents(folderId) {
  return api.get(`/api/v1/folders/${folderId}/contents`)
}

// 重命名文件夹
export function renameFolder(folderId, name) {
  return api.put(`/api/v1/folders/${folderId}`, { name })
}

// 移动文件夹
export function moveFolder(folderId, parentId) {
  return api.put(`/api/v1/folders/${folderId}/move`, { parent_id: parentId })
}

// 删除文件夹
export function deleteFolder(folderId) {
  return api.delete(`/api/v1/folders/${folderId}`)
}

// 获取文件夹访问链接
export function getFolderViewUrl(folderId) {
  return api.get(`/api/v1/folders/${folderId}/view`)
}

// 移动文件
export function moveFile(fileId, folderId) {
  return api.put(`/api/v1/files/${fileId}/move`, { folder_id: folderId })
}

// 获取文件访问链接
export function getFileViewUrl(fileId) {
  return api.get(`/api/v1/files/${fileId}/view`)
}
