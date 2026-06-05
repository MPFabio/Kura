import { apiClient } from './api'

// Types pour les jobs Ansible
export interface AnsibleJobSummary {
  id: number
  name: string
  status: string
  started?: string
  finished?: string
  elapsed?: number
  job_template?: number
  job_template_name?: string
  inventory?: number
  inventory_name?: string
  created_by?: number
  created_by_username?: string
}

export interface AnsibleJobDetail {
  id: number
  name: string
  status: string
  started?: string
  finished?: string
  elapsed?: number
  job_template?: number
  job_template_name?: string
  inventory?: number
  inventory_name?: string
  project?: number
  project_name?: string
  playbook?: string
  limit?: string
  verbosity?: number
  extra_vars?: Record<string, any>
  created_by?: number
  created_by_username?: string
  stdout?: string
}

export interface AnsibleJobResponse {
  items: AnsibleJobSummary[]
}

export interface AnsibleJobHistoryResponse {
  items: AnsibleJobSummary[]
  total: number
}

// Types pour les inventaires
export interface AnsibleInventorySummary {
  id: number
  name: string
  description?: string
  organization?: number
  organization_name?: string
  kind?: string
  host_count?: number
  created?: string
  modified?: string
}

export interface AnsibleInventoryDetail {
  id: number
  name: string
  description?: string
  organization?: number
  organization_name?: string
  kind?: string
  host_count?: number
  created?: string
  modified?: string
  variables?: Record<string, any>
}

export interface AnsibleInventoryResponse {
  items: AnsibleInventorySummary[]
}

// Types pour les hosts
export interface AnsibleHost {
  id: number
  name: string
  description?: string
  inventory: number
  enabled: boolean
  variables?: Record<string, any>
  created?: string
  modified?: string
}

export interface AnsibleHostResponse {
  items: AnsibleHost[]
}

// Types pour les templates de jobs (AWX summary_fields optionnel)
export interface AnsibleJobTemplateSummary {
  id: number
  name: string
  description?: string
  job_type?: string
  inventory?: number
  inventory_name?: string
  project?: number
  project_name?: string
  playbook?: string
  created?: string
  modified?: string
  summary_fields?: { inventory?: { name?: string }; project?: { name?: string } }
}

export interface AnsibleJobTemplateDetail {
  id: number
  name: string
  description?: string
  job_type?: string
  inventory?: number
  inventory_name?: string
  project?: number
  project_name?: string
  playbook?: string
  verbosity?: number
  limit?: string
  extra_vars?: Record<string, any>
  enabled?: boolean
  ask_variables_on_launch?: boolean
  ask_limit_on_launch?: boolean
  ask_tags_on_launch?: boolean
  ask_skip_tags_on_launch?: boolean
  ask_job_type_on_launch?: boolean
  ask_inventory_on_launch?: boolean
  ask_credential_on_launch?: boolean
  created?: string
  modified?: string
}

export interface AnsibleJobTemplateResponse {
  items: AnsibleJobTemplateSummary[]
}

export interface AnsibleJobLaunchRequest {
  extra_vars?: Record<string, any>
  limit?: string
  tags?: string
  skip_tags?: string
  verbosity?: number
  inventory?: number
  credentials?: number[]
}

export interface AnsibleJobLaunchResponse {
  job: number
  status: string
  message?: string
}

// Types pour les projets
export interface AnsibleProjectSummary {
  id: number
  name: string
  description?: string
  scm_type?: string
  scm_url?: string
  organization?: number
  organization_name?: string
  created?: string
  modified?: string
}

export interface AnsibleProjectDetail {
  id: number
  name: string
  description?: string
  scm_type?: string
  scm_url?: string
  scm_branch?: string
  scm_clean?: boolean
  scm_delete_on_update?: boolean
  organization?: number
  organization_name?: string
  created?: string
  modified?: string
}

export interface AnsibleProjectResponse {
  items: AnsibleProjectSummary[]
}

export interface AnsiblePlaybookResponse {
  items: string[]
}

// Types pour les credentials
export interface AnsibleCredentialSummary {
  id: number
  name: string
  description?: string
  credential_type?: number
  credential_type_name?: string
  organization?: number
  organization_name?: string
  created?: string
  modified?: string
}

export interface AnsibleCredentialDetail {
  id: number
  name: string
  description?: string
  credential_type: number
  credential_type_name?: string
  organization?: number
  organization_name?: string
  inputs?: Record<string, any>
  created?: string
  modified?: string
}

export interface AnsibleCredentialResponse {
  items: AnsibleCredentialSummary[]
}

export interface AnsibleCredentialCreate {
  name: string
  description?: string
  credential_type: number
  organization?: number
  inputs?: Record<string, any>
}

export interface AnsibleCredentialUpdate {
  name?: string
  description?: string
  credential_type?: number
  organization?: number
  inputs?: Record<string, any>
}

// Types pour les organisations
export interface AnsibleOrganizationSummary {
  id: number
  name: string
  description?: string
  created?: string
  modified?: string
}

export interface AnsibleOrganizationDetail {
  id: number
  name: string
  description?: string
  max_hosts?: number
  custom_virtualenv?: string
  default_environment?: number
  created?: string
  modified?: string
}

export interface AnsibleOrganizationResponse {
  items: AnsibleOrganizationSummary[]
}

export interface AnsibleOrganizationCreate {
  name: string
  description?: string
  max_hosts?: number
  custom_virtualenv?: string
  default_environment?: number
}

export interface AnsibleOrganizationUpdate {
  name?: string
  description?: string
  max_hosts?: number
  custom_virtualenv?: string
  default_environment?: number
}

// Types pour l'analyse de playbook
export interface AnsiblePlaybookAnalysisRequest {
  playbook_content: string
}

export interface AnsiblePlaybookAnalysis {
  parsed: {
    plays: Array<{
      name?: string
      hosts: string | string[]
      gather_facts?: boolean
      vars?: Record<string, any>
      vars_files?: string[]
      vars_prompt?: Array<{ name: string; prompt?: string; default?: any; private?: boolean }>
      pre_tasks?: any[]
      tasks: Array<{
        name?: string
        module?: string
        args?: Record<string, any>
        when?: string | string[]
        loop?: any
        until?: string
        retries?: number
        delay?: number
        register?: string
        changed_when?: string
        failed_when?: string
        ignore_errors?: boolean
        async_val?: number
        poll?: number
        delegate_to?: string
        delegate_facts?: boolean
        become?: boolean
        become_user?: string
        environment?: Record<string, string>
        tags?: string[]
        notify?: string[]
        type?: string
      }>
      handlers?: any[]
      post_tasks?: any[]
      roles?: string[]
      collections?: string[]
      become?: boolean
      become_user?: string
      become_method?: string
      environment?: Record<string, string>
    }>
    play_count: number
  }
  statistics: {
    total_plays: number
    total_tasks: number
    total_handlers: number
    total_pre_tasks: number
    total_post_tasks: number
    total_roles: number
    modules_used: string[]
    hosts_targeted: string[]
    become_used: boolean
    collections_used: string[]
  }
}

export const ansibleService = {
  // Jobs
  getJobs: async (): Promise<AnsibleJobResponse> => {
    try {
      const response = await apiClient.get<AnsibleJobResponse & { results?: AnsibleJobSummary[] }>('/v1/ansible/jobs')
      if (!response.data) return { items: [] }
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error('Erreur lors de la récupération des jobs:', error)
      throw error
    }
  },

  getJob: async (jobId: number, includeStdout = false): Promise<AnsibleJobDetail> => {
    try {
      const params = includeStdout ? '?include_stdout=true' : ''
      const response = await apiClient.get<AnsibleJobDetail>(`/v1/ansible/jobs/${jobId}${params}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du job ${jobId}:`, error)
      throw error
    }
  },

  getJobHistory: async (pageSize?: number): Promise<AnsibleJobHistoryResponse> => {
    try {
      const params = pageSize ? `?page_size=${pageSize}` : ''
      const response = await apiClient.get<
        AnsibleJobHistoryResponse & { results?: AnsibleJobSummary[]; count?: number }
      >(`/v1/ansible/jobs/history${params}`)
      if (!response.data) return { items: [], total: 0 }
      const items = response.data.items ?? response.data.results ?? []
      const total = response.data.total ?? response.data.count ?? items.length
      return { items, total }
    } catch (error) {
      console.error('Erreur lors de la récupération de l\'historique des jobs:', error)
      throw error
    }
  },

  // Inventaires
  getInventories: async (): Promise<AnsibleInventoryResponse> => {
    try {
      const response = await apiClient.get<
        AnsibleInventoryResponse & { results?: AnsibleInventorySummary[] }
      >('/v1/ansible/inventories')
      if (!response.data) return { items: [] }
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error('Erreur lors de la récupération des inventaires:', error)
      throw error
    }
  },

  getInventory: async (inventoryId: number): Promise<AnsibleInventoryDetail> => {
    try {
      const response = await apiClient.get<AnsibleInventoryDetail>(`/v1/ansible/inventories/${inventoryId}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération de l'inventaire ${inventoryId}:`, error)
      throw error
    }
  },

  getInventoryHosts: async (inventoryId: number): Promise<AnsibleHostResponse> => {
    try {
      const response = await apiClient.get<AnsibleHostResponse & { results?: AnsibleHost[] }>(
        `/v1/ansible/inventories/${inventoryId}/hosts`
      )
      if (!response.data) return { items: [] }
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error(`Erreur lors de la récupération des hosts de l'inventaire ${inventoryId}:`, error)
      throw error
    }
  },

  // Templates de jobs
  getJobTemplates: async (): Promise<AnsibleJobTemplateResponse> => {
    try {
      const response = await apiClient.get<AnsibleJobTemplateResponse & { results?: AnsibleJobTemplateSummary[] }>(
        '/v1/ansible/job-templates'
      )
      if (!response.data) return { items: [] }
      // AWX retourne "results", le frontend attend "items"
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error('Erreur lors de la récupération des templates de jobs:', error)
      throw error
    }
  },

  getJobTemplate: async (templateId: number): Promise<AnsibleJobTemplateDetail> => {
    try {
      const response = await apiClient.get<AnsibleJobTemplateDetail>(`/v1/ansible/job-templates/${templateId}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du template ${templateId}:`, error)
      throw error
    }
  },

  launchJobTemplate: async (templateId: number, launchData?: AnsibleJobLaunchRequest): Promise<AnsibleJobLaunchResponse> => {
    try {
      const response = await apiClient.post<AnsibleJobLaunchResponse>(
        `/v1/ansible/job-templates/${templateId}/launch`,
        launchData || {}
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors du lancement du template ${templateId}:`, error)
      throw error
    }
  },

  // Projets
  getProjects: async (): Promise<AnsibleProjectResponse> => {
    try {
      const response = await apiClient.get<AnsibleProjectResponse & { results?: AnsibleProjectSummary[] }>(
        '/v1/ansible/projects'
      )
      if (!response.data) return { items: [] }
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error('Erreur lors de la récupération des projets:', error)
      throw error
    }
  },

  getProject: async (projectId: number): Promise<AnsibleProjectDetail> => {
    try {
      const response = await apiClient.get<AnsibleProjectDetail>(`/v1/ansible/projects/${projectId}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du projet ${projectId}:`, error)
      throw error
    }
  },

  getProjectPlaybooks: async (projectId: number): Promise<AnsiblePlaybookResponse> => {
    try {
      const response = await apiClient.get<AnsiblePlaybookResponse & { results?: string[] }>(
        `/v1/ansible/projects/${projectId}/playbooks`
      )
      if (!response.data) return { items: [] }
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error(`Erreur lors de la récupération des playbooks du projet ${projectId}:`, error)
      throw error
    }
  },

  // Credentials
  getCredentials: async (): Promise<AnsibleCredentialResponse> => {
    try {
      const response = await apiClient.get<
        AnsibleCredentialResponse & { results?: AnsibleCredentialSummary[] }
      >('/v1/ansible/credentials')
      if (!response.data) return { items: [] }
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error('Erreur lors de la récupération des credentials:', error)
      throw error
    }
  },

  getCredential: async (credentialId: number): Promise<AnsibleCredentialDetail> => {
    try {
      const response = await apiClient.get<AnsibleCredentialDetail>(`/v1/ansible/credentials/${credentialId}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du credential ${credentialId}:`, error)
      throw error
    }
  },

  createCredential: async (credential: AnsibleCredentialCreate): Promise<AnsibleCredentialDetail> => {
    try {
      const response = await apiClient.post<AnsibleCredentialDetail>('/v1/ansible/credentials', credential)
      return response.data
    } catch (error) {
      console.error('Erreur lors de la création du credential:', error)
      throw error
    }
  },

  updateCredential: async (credentialId: number, credential: AnsibleCredentialUpdate): Promise<AnsibleCredentialDetail> => {
    try {
      const response = await apiClient.put<AnsibleCredentialDetail>(`/v1/ansible/credentials/${credentialId}`, credential)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la mise à jour du credential ${credentialId}:`, error)
      throw error
    }
  },

  deleteCredential: async (credentialId: number): Promise<void> => {
    try {
      await apiClient.delete(`/v1/ansible/credentials/${credentialId}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression du credential ${credentialId}:`, error)
      throw error
    }
  },

  // Organizations
  getOrganizations: async (): Promise<AnsibleOrganizationResponse> => {
    try {
      const response = await apiClient.get<
        AnsibleOrganizationResponse & { results?: AnsibleOrganizationSummary[] }
      >('/v1/ansible/organizations')
      if (!response.data) return { items: [] }
      const items = response.data.items ?? response.data.results ?? []
      return { items }
    } catch (error) {
      console.error('Erreur lors de la récupération des organisations:', error)
      throw error
    }
  },

  getOrganization: async (organizationId: number): Promise<AnsibleOrganizationDetail> => {
    try {
      const response = await apiClient.get<AnsibleOrganizationDetail>(`/v1/ansible/organizations/${organizationId}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération de l'organisation ${organizationId}:`, error)
      throw error
    }
  },

  createOrganization: async (organization: AnsibleOrganizationCreate): Promise<AnsibleOrganizationDetail> => {
    try {
      const response = await apiClient.post<AnsibleOrganizationDetail>('/v1/ansible/organizations', organization)
      return response.data
    } catch (error) {
      console.error('Erreur lors de la création de l\'organisation:', error)
      throw error
    }
  },

  updateOrganization: async (organizationId: number, organization: AnsibleOrganizationUpdate): Promise<AnsibleOrganizationDetail> => {
    try {
      const response = await apiClient.put<AnsibleOrganizationDetail>(`/v1/ansible/organizations/${organizationId}`, organization)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la mise à jour de l'organisation ${organizationId}:`, error)
      throw error
    }
  },

  deleteOrganization: async (organizationId: number): Promise<void> => {
    try {
      await apiClient.delete(`/v1/ansible/organizations/${organizationId}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression de l'organisation ${organizationId}:`, error)
      throw error
    }
  },

  // Configuration Semaphore
  getConfig: async (): Promise<{
    semaphore_url: string
    semaphore_project_id: number
    has_token: boolean
    configured: boolean
  }> => {
    const response = await apiClient.get('/v1/ansible/config')
    return response.data
  },

  setConfig: async (data: {
    semaphore_url?: string
    token?: string
    semaphore_project_id?: number
  }): Promise<{ configured: boolean }> => {
    const response = await apiClient.post('/v1/ansible/config', data)
    return response.data
  },

  // Analyse de playbook
  analyzePlaybook: async (playbookContent: string): Promise<AnsiblePlaybookAnalysis> => {
    try {
      const response = await apiClient.post<AnsiblePlaybookAnalysis>('/v1/ansible/playbooks/analyze', {
        playbook_content: playbookContent,
      })
      return response.data
    } catch (error) {
      console.error('Erreur lors de l\'analyse du playbook:', error)
      throw error
    }
  },
}
