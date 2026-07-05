/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
export type ApiResponse<T> = {
  success: boolean
  message?: string
  data?: T
}

export type PagePayload<T> = {
  page?: number
  page_size?: number
  total?: number
  items?: T[]
}

export type PrivacyRequestType = 'access' | 'correction' | 'deletion'

export type PrivacyRequestStatus =
  | 'pending'
  | 'processing'
  | 'completed'
  | 'rejected'
  | 'cancelled'

export type PublicFeedbackType = 'complaint' | 'feedback' | 'other'

export type PublicFeedbackStatus =
  | 'pending'
  | 'processing'
  | 'resolved'
  | 'closed'
  | 'rejected'

export type PersonalInfoSnapshot = {
  id: number
  username: string
  display_name?: string
  role?: number
  status?: number
  email?: string
  phone_number?: string
  github_id?: string
  discord_id?: string
  oidc_id?: string
  wechat_id?: string
  telegram_id?: string
  linux_do_id?: string
  group?: string
  quota?: number
  used_quota?: number
  request_count?: number
  aff_code?: string
  aff_count?: number
  aff_quota?: number
  aff_history_quota?: number
  inviter_id?: number
  stripe_customer?: string
  created_at?: number
  last_login_at?: number
}

export type PrivacyRequest = {
  id: number
  user_id: number
  username: string
  contact_name: string
  contact_email: string
  contact_phone: string
  request_type: PrivacyRequestType
  content: string
  status: PrivacyRequestStatus
  admin_id?: number
  admin_name?: string
  admin_note?: string
  execute_account_deletion?: boolean
  created_at?: number
  updated_at?: number
  handled_at?: number
}

export type PublicFeedback = {
  id: number
  user_id: number
  username: string
  contact_name: string
  contact_email: string
  contact_phone: string
  feedback_type: PublicFeedbackType
  title: string
  content: string
  status: PublicFeedbackStatus
  tracking_code: string
  admin_id?: number
  admin_name?: string
  admin_note?: string
  created_at?: number
  updated_at?: number
  handled_at?: number
}

export type CreatePrivacyRequestPayload = {
  request_type: PrivacyRequestType
  contact_name: string
  contact_email: string
  contact_phone: string
  content: string
}

export type CreatePublicFeedbackPayload = {
  feedback_type: PublicFeedbackType
  contact_name: string
  contact_email: string
  contact_phone: string
  title: string
  content: string
}

export type AdminPrivacyRequestUpdatePayload = {
  status: PrivacyRequestStatus
  admin_note: string
  execute_account_deletion: boolean
}

export type AdminPublicFeedbackUpdatePayload = {
  status: PublicFeedbackStatus
  admin_note: string
}

export type ComplianceListParams = {
  p?: number
  page_size?: number
  status?: string
  request_type?: string
  feedback_type?: string
}
