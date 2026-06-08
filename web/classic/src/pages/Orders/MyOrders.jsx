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

const defaultFilters = { keyword: '', status: '', order_type: '' };

const MyOrders = () => {
  const { t } = useTranslation();
  const [filters, setFilters] = useState(defaultFilters);
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detail, setDetail] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [refundVisible, setRefundVisible] = useState(false);
  const [refundLoading, setRefundLoading] = useState(false);
  const [activeOrder, setActiveOrder] = useState(null);

  const loadOrders = async () => {
    setLoading(true);
    try {
      const res = await paymentOrdersApi.listMyOrders({
        p: page,
        page_size: pageSize,
        ...filters,
      });
      if (res.data.success) {
        setOrders(res.data.data.items || []);
        setTotal(res.data.data.total || 0);
      } else {
        Toast.error({ content: res.data.message || t('加载失败') });
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders();
  }, [page, pageSize, filters]);

  const updateFilter = (key, value) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setPage(1);
  };

  const openDetail = async (order) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await paymentOrdersApi.getMyOrder(order.order_type, order.id);
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
        const res = await paymentOrdersApi.cancelMyOrder(order.order_type, order.id);
        if (res.data.success) {
          Toast.success({ content: t('操作成功') });
          loadOrders();
        } else {
          Toast.error({ content: res.data.message || t('操作失败') });
        }
      },
    });
  };

  const submitRefund = async (values) => {
    setRefundLoading(true);
    try {
      const res = await paymentOrdersApi.requestMyRefund(activeOrder.order_type, activeOrder.id, {
        reason: values.reason,
      });
      if (res.data.success) {
        Toast.success({ content: t('退款申请已提交') });
        setRefundVisible(false);
        loadOrders();
      } else {
        Toast.error({ content: res.data.message || t('操作失败') });
      }
    } finally {
      setRefundLoading(false);
    }
  };

  return (
    <div className='mt-[60px] px-2'>
      <Card title={t('我的订单')}>
        <Space wrap className='mb-4'>
          <Input
            style={{ width: 220 }}
            placeholder={t('搜索订单号')}
            value={filters.keyword}
            onChange={(value) => updateFilter('keyword', value)}
          />
          <Select
            style={{ width: 150 }}
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
          <Button onClick={() => setFilters(defaultFilters)}>{t('重置')}</Button>
        </Space>
        <PaymentOrderTable
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
          onRefund={(order) => {
            setActiveOrder(order);
            setRefundVisible(true);
          }}
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
        mode='user'
        loading={refundLoading}
        onCancel={() => setRefundVisible(false)}
        onSubmit={submitRefund}
        t={t}
      />
    </div>
  );
};

export default MyOrders;
