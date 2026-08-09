<?php
declare(strict_types=1);

final class Paginator
{
    public function __construct(
        private readonly int $currentPage,
        private readonly int $perPage,
        private readonly int $total
    ) {}

    public function total(): int { return $this->total; }

    public function totalPages(): int
    {
        return max(1, (int) ceil($this->total / max(1, $this->perPage)));
    }

    public function currentPage(): int
    {
        return min(max(1, $this->currentPage), $this->totalPages());
    }

    public function from(): int
    {
        if ($this->total === 0) return 0;
        return (($this->currentPage() - 1) * $this->perPage) + 1;
    }

    public function to(): int
    {
        return min($this->currentPage() * $this->perPage, $this->total);
    }

    public function hasPrevious(): bool { return $this->currentPage() > 1; }
    public function hasNext(): bool { return $this->currentPage() < $this->totalPages(); }

    public function pages(int $window = 2): array
    {
        $last = $this->totalPages();
        $current = $this->currentPage();

        if ($last <= ($window * 2) + 5) return range(1, $last);

        $pages = [1];
        $start = max(2, $current - $window);
        $end = min($last - 1, $current + $window);

        if ($start > 2) $pages[] = null;
        for ($i = $start; $i <= $end; $i++) $pages[] = $i;
        if ($end < $last - 1) $pages[] = null;

        $pages[] = $last;
        return $pages;
    }
}
