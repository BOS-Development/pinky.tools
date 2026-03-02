import { render, screen, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import PendingSales from '../PendingSales';

jest.mock('next-auth/react');

const makeSale = (overrides: Record<string, unknown> = {}) => ({
  id: 1,
  forSaleItemId: 100,
  buyerUserId: 10,
  buyerName: 'Buyer One',
  sellerUserId: 20,
  typeId: 1000,
  typeName: 'Tritanium',
  locationId: 60003760,
  locationName: 'Jita IV - Moon 4',
  quantityPurchased: 100,
  pricePerUnit: 5,
  totalPrice: 500,
  status: 'pending',
  isAutoFulfilled: false,
  purchasedAt: '2025-01-15T10:00:00Z',
  ...overrides,
});

describe('PendingSales', () => {
  const mockSession = {
    data: { user: { name: 'Test User' }, providerAccountId: '123456' },
    status: 'authenticated',
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (useSession as jest.Mock).mockReturnValue(mockSession);
  });

  it('shows empty state when no pending sales', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(<PendingSales />);

    await waitFor(() => {
      expect(screen.getByText('No pending sales')).toBeInTheDocument();
    });
  });

  it('aggregates items with the same typeId into one row', async () => {
    const sales = [
      makeSale({ id: 1, typeId: 1000, typeName: 'Tritanium', quantityPurchased: 100, pricePerUnit: 5, totalPrice: 500 }),
      makeSale({ id: 2, typeId: 1000, typeName: 'Tritanium', quantityPurchased: 200, pricePerUnit: 5, totalPrice: 1000 }),
      makeSale({ id: 3, typeId: 2000, typeName: 'Pyerite', quantityPurchased: 50, pricePerUnit: 10, totalPrice: 500 }),
    ];

    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => sales,
    });

    render(<PendingSales />);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });

    // Tritanium should appear once (aggregated), not twice
    const tritaniumCells = screen.getAllByText('Tritanium');
    expect(tritaniumCells).toHaveLength(1);

    // Pyerite should appear once
    const pyeriteCells = screen.getAllByText('Pyerite');
    expect(pyeriteCells).toHaveLength(1);

    // Aggregated quantity for Tritanium: 100 + 200 = 300
    expect(screen.getByText('300')).toBeInTheDocument();

    // Pyerite quantity: 50
    expect(screen.getByText('50')).toBeInTheDocument();

    // Aggregated total for Tritanium: 500 + 1000 = 1,500 ISK
    expect(screen.getByText('1,500 ISK')).toBeInTheDocument();
  });

  it('computes weighted average price per unit for aggregated items', async () => {
    // Two sales of same type at different prices
    const sales = [
      makeSale({ id: 1, typeId: 1000, typeName: 'Tritanium', quantityPurchased: 100, pricePerUnit: 4, totalPrice: 400 }),
      makeSale({ id: 2, typeId: 1000, typeName: 'Tritanium', quantityPurchased: 100, pricePerUnit: 6, totalPrice: 600 }),
    ];

    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => sales,
    });

    render(<PendingSales />);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });

    // Weighted average: (400 + 600) / (100 + 100) = 5
    expect(screen.getByText('5 ISK')).toBeInTheDocument();

    // Total quantity: 200
    expect(screen.getByText('200')).toBeInTheDocument();
  });

  it('shows Auto chip when any aggregated sale is auto-fulfilled', async () => {
    const sales = [
      makeSale({ id: 1, typeId: 1000, typeName: 'Tritanium', isAutoFulfilled: false }),
      makeSale({ id: 2, typeId: 1000, typeName: 'Tritanium', isAutoFulfilled: true }),
    ];

    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => sales,
    });

    render(<PendingSales />);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });

    expect(screen.getByText('Auto')).toBeInTheDocument();
  });

  it('does not show Requested column (removed in aggregation)', async () => {
    const sales = [makeSale()];

    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => sales,
    });

    render(<PendingSales />);

    await waitFor(() => {
      expect(screen.getByText('Item')).toBeInTheDocument();
    });

    expect(screen.queryByText('Requested')).not.toBeInTheDocument();
  });

  it('joins transaction notes from multiple sales of same type', async () => {
    const sales = [
      makeSale({ id: 1, typeId: 1000, typeName: 'Tritanium', transactionNotes: 'urgent' }),
      makeSale({ id: 2, typeId: 1000, typeName: 'Tritanium', transactionNotes: 'priority' }),
    ];

    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => sales,
    });

    render(<PendingSales />);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });

    expect(screen.getByText('urgent; priority')).toBeInTheDocument();
  });

  it('displays correct aggregated item count in accordion header', async () => {
    const sales = [
      makeSale({ id: 1, typeId: 1000, typeName: 'Tritanium' }),
      makeSale({ id: 2, typeId: 1000, typeName: 'Tritanium' }),
      makeSale({ id: 3, typeId: 2000, typeName: 'Pyerite' }),
    ];

    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => sales,
    });

    render(<PendingSales />);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });

    // 2 unique item types = "2 items"
    expect(screen.getByText('2 items')).toBeInTheDocument();
  });
});
