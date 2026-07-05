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
import { api } from '@/lib/api'
import type {
  AdminPrivacyRequestUpdatePayload,
  AdminPublicFeedbackUpdatePayload,
  ApiResponse,
  ComplianceListParams,
  CreatePrivacyRequestPayload,
  CreatePublicFeedbackPayload,
  PagePayload,
  PersonalInfoSnapshot,
  PrivacyRequest,
  PublicFeedback,
} from './types'

export async function getPersonalInfoSnapshot(): Promise<
  ApiResponse<PersonalInfoSnapshot>
> {
  const res = await api.get('/api/privacy/personal-info')
  return res.data
}

export async function listMyPrivacyRequests(
  params: ComplianceListParams = {}
): Promise<ApiResponse<PagePayload<PrivacyRequest>>> {
  const res = await api.get('/api/privacy/requests', { params })
  return res.data
}

export async function createPrivacyRequest(
  payload: CreatePrivacyRequestPayload
): Promise<ApiResponse<PrivacyRequest>> {
  const res = await api.post('/api/privacy/requests', payload)
  return res.data
}

export async function cancelPrivacyRequest(
  id: number
): Promise<ApiResponse<PrivacyRequest>> {
  const res = await api.post(`/api/privacy/requests/${id}/cancel`)
  return res.data
}

export async function createPublicFeedback(
  payload: CreatePublicFeedbackPayload,
  turnstileToken?: string
): Promise<ApiResponse<{ id: number; tracking_code: string }>> {
  const res = await api.post('/api/feedback', payload, {
    params: turnstileToken ? { turnstile: turnstileToken } : undefined,
  })
  return res.data
}

export async function listAdminPrivacyRequests(
  params: ComplianceListParams = {}
): Promise<ApiResponse<PagePayload<PrivacyRequest>>> {
  const res = await api.get('/api/privacy/admin/requests', {
    params,
    disableDuplicate: true,
  })
  return res.data
}

export async function getAdminPrivacyRequest(
  id: number
): Promise<ApiResponse<PrivacyRequest>> {
  const res = await api.get(`/api/privacy/admin/requests/${id}`, {
    disableDuplicate: true,
  })
  return res.data
}

export async function updateAdminPrivacyRequest(
  id: number,
  payload: AdminPrivacyRequestUpdatePayload
): Promise<ApiResponse<PrivacyRequest>> {
  const res = await api.patch(`/api/privacy/admin/requests/${id}`, payload)
  return res.data
}

export async function listAdminFeedback(
  params: ComplianceListParams = {}
): Promise<ApiResponse<PagePayload<PublicFeedback>>> {
  const res = await api.get('/api/feedback/admin', {
    params,
    disableDuplicate: true,
  })
  return res.data
}

export async function getAdminFeedback(
  id: number
): Promise<ApiResponse<PublicFeedback>> {
  const res = await api.get(`/api/feedback/admin/${id}`, {
    disableDuplicate: true,
  })
  return res.data
}

export async function updateAdminFeedback(
  id: number,
  payload: AdminPublicFeedbackUpdatePayload
): Promise<ApiResponse<PublicFeedback>> {
  const res = await api.patch(`/api/feedback/admin/${id}`, payload)
  return res.data
}
