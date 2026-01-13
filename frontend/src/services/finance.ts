import api from './api'

export interface Account {
  id: string
  league_id: string
  owner_id: string
  owner_type: 'team' | 'participant' | 'system'
  owner_name: string
  balance: number
  created_at: string
  updated_at: string
}

export interface Transaction {
  id: string
  league_id: string
  from_account_id: string
  to_account_id: string
  from_name: string
  to_name: string
  amount: number
  category: 'prize' | 'transfer' | 'penalty' | 'sponsorship' | 'other'
  description?: string
  created_at: string
}

export interface ListAccountsResponse {
  accounts: Account[]
  total: number
}

export interface ListTransactionsResponse {
  transactions: Transaction[]
  total: number
  page: number
  total_pages: number
}

export interface TeamBalance {
  team_id: string
  team_name: string
  balance: number
}

export interface MonthlyFlow {
  month: string
  income: number
  expense: number
}

export interface FinanceStats {
  total_circulation: number
  team_balances: TeamBalance[]
  category_totals: Record<string, number>
  monthly_flow: MonthlyFlow[]
}

export const financeService = {
  // Accounts
  listAccounts: async (leagueId: string) => {
    const response = await api.get<ListAccountsResponse>(`/leagues/${leagueId}/accounts`)
    return response.data
  },

  getAccount: async (accountId: string) => {
    const response = await api.get<Account>(`/accounts/${accountId}`)
    return response.data
  },

  setBalance: async (accountId: string, balance: number) => {
    const response = await api.put<Account>(`/admin/accounts/${accountId}/balance`, { balance })
    return response.data
  },

  // Transactions
  createTransaction: async (leagueId: string, data: {
    from_account_id: string
    to_account_id: string
    amount: number
    category: string
    description?: string
  }) => {
    const response = await api.post<Transaction>(`/admin/leagues/${leagueId}/transactions`, data)
    return response.data
  },

  listTransactions: async (leagueId: string, params?: {
    page?: number
    limit?: number
    account_id?: string
    category?: string
  }) => {
    const response = await api.get<ListTransactionsResponse>(`/leagues/${leagueId}/transactions`, { params })
    return response.data
  },

  getAccountTransactions: async (accountId: string, params?: {
    page?: number
    limit?: number
  }) => {
    const response = await api.get<{ transactions: Transaction[], total: number, balance: number }>(
      `/accounts/${accountId}/transactions`, { params }
    )
    return response.data
  },

  // Stats
  getFinanceStats: async (leagueId: string) => {
    const response = await api.get<FinanceStats>(`/leagues/${leagueId}/finance/stats`)
    return response.data
  },
}

export default financeService
