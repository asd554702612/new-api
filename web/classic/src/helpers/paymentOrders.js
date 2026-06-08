/*
Copyright (C) 2025 QuantumNous

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

import { API } from './api';

export const PAYMENT_ORDER_TYPES = {
  balance: 'balance',
  subscription: 'subscription',
};

export const PAYMENT_ORDER_STATUSES = [
  'PENDING',
  'COMPLETED',
  'FAILED',
  'EXPIRED',
  'CANCELLED',
  'REFUND_REQUESTED',
  'REFUNDED',
  'REFUND_FAILED',
];

const buildParams = (params = {}) => {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      search.set(key, value);
    }
  });
  const query = search.toString();
  return query ? `?${query}` : '';
};

export const paymentOrdersApi = {
  listAdminOrders: (params) =>
    API.get(`/api/payment/admin/orders${buildParams(params)}`),
  listMyOrders: (params) => API.get(`/api/payment/orders/my${buildParams(params)}`),
  getAdminOrder: (orderType, id) =>
    API.get(`/api/payment/admin/orders/${orderType}/${id}`),
  getMyOrder: (orderType, id) => API.get(`/api/payment/orders/${orderType}/${id}`),
  cancelAdminOrder: (orderType, id) =>
    API.post(`/api/payment/admin/orders/${orderType}/${id}/cancel`),
  cancelMyOrder: (orderType, id) =>
    API.post(`/api/payment/orders/${orderType}/${id}/cancel`),
  retryAdminOrder: (orderType, id) =>
    API.post(`/api/payment/admin/orders/${orderType}/${id}/retry`),
  refundAdminOrder: (orderType, id, payload) =>
    API.post(`/api/payment/admin/orders/${orderType}/${id}/refund`, payload),
  requestMyRefund: (orderType, id, payload) =>
    API.post(`/api/payment/orders/${orderType}/${id}/refund-request`, payload),
  dashboard: (days = 30) =>
    API.get(`/api/payment/admin/dashboard${buildParams({ days })}`),
  getActivityConfig: (activityType) =>
    API.get(`/api/payment/admin/activities/${activityType}/config`),
  updateActivityConfig: (activityType, payload) =>
    API.put(`/api/payment/admin/activities/${activityType}/config`, payload),
  getActivityStats: (activityType) =>
    API.get(`/api/payment/admin/activities/${activityType}/stats`),
  getLuckyWheelSummary: () => API.get('/api/payment/lucky-wheel'),
  drawLuckyWheel: (sessionId) =>
    API.post('/api/payment/lucky-wheel/draw', { session_id: sessionId }),
  getRechargeActivitySummary: () => API.get('/api/payment/recharge-activity'),
  drawRechargeActivity: (chanceId) =>
    API.post('/api/payment/recharge-activity/draw', { chance_id: chanceId }),
  getLuckyWheelConfig: () => API.get('/api/payment/admin/lucky-wheel/config'),
  updateLuckyWheelConfig: (payload) =>
    API.put('/api/payment/admin/lucky-wheel/config', payload),
  getLuckyWheelStats: () => API.get('/api/payment/admin/lucky-wheel/stats'),
  getRechargeActivityConfig: () =>
    API.get('/api/payment/admin/recharge-activity/config'),
  updateRechargeActivityConfig: (payload) =>
    API.put('/api/payment/admin/recharge-activity/config', payload),
  getRechargeActivityStats: (params) =>
    API.get(`/api/payment/admin/recharge-activity/stats${buildParams(params)}`),
  updateRechargeActivityFulfillment: (id, payload) =>
    API.put(`/api/payment/admin/recharge-activity/records/${id}/fulfillment`, payload),
};
