import { apiClient } from './api'

export interface RepoTreeEntry {
  name: string
  path: string
  type: 'file' | 'dir'
  size?: number
}

export interface FileContent {
  path: string
  content: string
  size: number
  truncated?: boolean
}

export interface CommitSummary {
  sha: string
  message: string
  author: string
  date: string
  url: string
}

export interface CommitFile {
  filename: string
  status: string
  additions: number
  deletions: number
  patch?: string
}

export interface CommitDetail extends CommitSummary {
  files: CommitFile[]
}

export interface CodeRepository {
  mapping_id: string
  full_name: string
}

export const codeService = {
  listRepositories: async (projectId: string): Promise<CodeRepository[]> => {
    const response = await apiClient.get<{ items: CodeRepository[] }>('/v1/code/repos', {
      params: { project_id: projectId },
    })
    return response.data?.items ?? []
  },

  getTree: async (repo: string, path = '', ref = ''): Promise<RepoTreeEntry[]> => {
    const response = await apiClient.get<RepoTreeEntry[]>('/v1/code/tree', { params: { repo, path, ref } })
    return response.data ?? []
  },

  getFile: async (repo: string, path: string, ref = ''): Promise<FileContent> => {
    const response = await apiClient.get<FileContent>('/v1/code/file', { params: { repo, path, ref } })
    return response.data
  },

  getCommits: async (repo: string, path = '', ref = '', page = 1): Promise<CommitSummary[]> => {
    const response = await apiClient.get<CommitSummary[]>('/v1/code/commits', { params: { repo, path, ref, page } })
    return response.data ?? []
  },

  getCommitDetail: async (repo: string, sha: string): Promise<CommitDetail> => {
    const response = await apiClient.get<CommitDetail>(`/v1/code/commits/${sha}`, { params: { repo } })
    return response.data
  },
}
