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

import React, { useEffect, useState } from 'react';
import { Button, Card, Input, Modal, Select, Space, Toast } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { paymentOrdersApi, PAYMENT_ORDER_STATUSES } from '../../helpers/paymentOrders';
import PaymentOrderTable from '../../components/payment-orders/PaymentOrderTable';
import PaymentOrderDetailModal from '../../components/payment-orders/PaymentOrderDetailModal';
import PaymentRefundModal from '../../components/payment-orders/PaymentRefundModal';

const defaultFilters = {
  keyword: '',
  status: '',
  order_type: '',
  payment_type: '',
  user_id: '',
};

const AdminOrders = () => {
  const { t } = useTranslation();
  const [filters, setFilters] = useState(defaultFilters);
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detail, setDetail] = useState(null);
  const [refundVisible, setRefundVisible] = useState(false);
  const [refundLoading, setRefundLoading] = useState(false);
  const [activeOrder, setActiveOrder] = useState(null);

  const loadOrders = async () => {
    setLoading(true);
    try {
      const res = await paymentOrdersApi.listAdminOrders({
        p: page,
        page_size: pageSize,
        ...filters,
      });
      const { success, message, data } = res.data;
      if (success) {
        setOrders(data.items || []);
        setTotal(data.total || 0);
      } else {
        Toast.error({ content: message || t('加载失败') });
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders();
  }, [page, pageSize, filters]);

  const openDetail = async (order) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await paymentOrdersApi.getAdminOrder(order.order_type, order.id);
      if (res.data.success) {
        setDetail(res.data.data);
      } else {
        Toast.error({ content: res.data.message || t('加载失败') });
      }
    } finally {
      setDetailLoading(false);
    }
  };

  const confirmCancel = (order) => {
    Modal.confirm({
      title: t('取消订单'),
      content: t('确认取消该待支付订单？'),
      onOk: async () => {
        const res = await paymentOrdersApi.cancelAdminOrder(order.order_type, order.id);
        if (res.data.success) {
          Toast.success({ content: t('操作成功') });
          loadOrders();
        } else {
          Toast.error({ content: res.data.message || t('操作失败') });
        }
      },
    });
  };

  const confirmRetry = (order) => {
    Modal.confirm({
      title: t('补发订单'),
      content: t('确认重试该订单的履约流程？'),
      onOk: async () => {
        const res = await paymentOrdersApi.retryAdminOrder(order.order_type, order.id);
        if (res.data.success) {
          Toast.success({ content: t('操作成功') });
          loadOrders();
        } else {
          Toast.error({ content: res.data.message || t('操作失败') });
        }
      },
    });
  };

  const openRefund = (order) => {
    setActiveOrder(order);
    setRefundVisible(true);
  };

  const submitRefund = async (values) => {
    setRefundLoading(true);
    try {
      const res = await paymentOrdersApi.refundAdminOrder(
        activeOrder.order_type,
        activeOrder.id,
        values,
      );
      if (res.data.success) {
        const refundNo = res.data.data?.provider_refund_no;
        Toast.success({
          content: refundNo
            ? `${t('退款已发起')}: ${refundNo}`
            : t('退款已发起'),
        });
        setRefundVisible(false);
        loadOrders();
      } else {
        Toast.error({ content: res.data.message || t('操作失败') });
      }
    } finally {
      setRefundLoading(false);
    }
  };

  const updateFilter = (key, value) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setPage(1);
  };

  return (
    <div className='mt-[60px] px-2'>
      <Card title={t('订单管理')}>
        <Space wrap className='mb-4'>
          <Input
            style={{ width: 220 }}
            placeholder={t('搜索订单号')}
            value={filters.keyword}
            onChange={(value) => updateFilter('keyword', value)}
          />
          <Input
            style={{ width: 140 }}
            placeholder={t('用户ID')}
            value={filters.user_id}
            onChange={(value) => updateFilter('user_id', value)}
          />
          <Select
            style={{ width: 150 }}
            placeholder={t('订单类型')}
            value={filters.order_type}
            onChange={(value) => updateFilter('order_type', value)}
            optionList={[
              { label: t('全部类型'), value: '' },
              { label: t('充值订单'), value: 'balance' },
              { label: t('订阅订单'), value: 'subscription' },
            ]}
          />
          <Select
            style={{ width: 160 }}
            placeholder={t('订单状态')}
            value={filters.status}
            onChange={(value) => updateFilter('status', value)}
            optionList={[
              { label: t('全部状态'), value: '' },
              ...PAYMENT_ORDER_STATUSES.map((status) => ({
                label: t(status),
                value: status,
              })),
            ]}
          />
          <Input
            style={{ width: 160 }}
            placeholder={t('支付方式')}
            value={filters.payment_type}
            onChange={(value) => updateFilter('payment_type', value)}
          />
          <Button onClick={() => setFilters(defaultFilters)}>{t('重置')}</Button>
        </Space>
        <PaymentOrderTable
          admin
          orders={orders}
          loading={loading}
          pagination={{ page, pageSize, total }}
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size);
            setPage(1);
          }}
          onDetail={openDetail}
          onCancel={confirmCancel}
          onRetry={confirmRetry}
          onRefund={openRefund}
          t={t}
        />
      </Card>
      <PaymentOrderDetailModal
        visible={detailVisible}
        detail={detail}
        loading={detailLoading}
        onCancel={() => setDetailVisible(false)}
        t={t}
      />
      <PaymentRefundModal
        visible={refundVisible}
        order={activeOrder}
        mode='admin'
        loading={refundLoading}
        onCancel={() => setRefundVisible(false)}
        onSubmit={submitRefund}
        t={t}
      />
    </div>
  );
};

export default AdminOrders;
